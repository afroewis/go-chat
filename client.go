package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

// Sends messages from the websocket connection to the hub.
// The application runs readLoop in a per-connection goroutine.
func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	// todo move to main.go?
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, jsonMessage, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if !Debug {
			// Publish to redis
			jsonMessage = bytes.TrimSpace(bytes.Replace(jsonMessage, newline, space, -1))
			dec := json.NewDecoder(strings.NewReader(string(jsonMessage)))

			var message Message

			if err := dec.Decode(&message); err != nil && err != io.EOF {
				log.Fatal(err)
			}

			// todo move to constant
			message.Hostname, _ = os.Hostname()

			updatedJsonMessage, _ := json.Marshal(message)

			publishErr := c.hub.pubConn.Conn.Send("PUBLISH", message.Chatroom, updatedJsonMessage)
			if publishErr != nil {
				log.Fatal("Error when publishing to redis: %v", publishErr)
				return
			}
			flushErr := c.hub.pubConn.Conn.Flush()
			if flushErr != nil {
				log.Fatal("Error when flushing to redis: %v", flushErr)
				return
			}

			log.Printf("Published: %s\n", jsonMessage)
		} else {
			// Publish to websocket directly
			c.hub.broadcast <- jsonMessage
		}
	}
}

// writeLoop sends messages from the hub to the websocket connection.
// A goroutine running writeLoop is started for each connection.
//goland:noinspection ALL
func (c *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writeLoop()
	go client.readLoop()
}
