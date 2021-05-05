package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client

	subConn *redis.PubSubConn

	pubConn *redis.PubSubConn
}

func newHub() *Hub {
	var hub = Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	if !Debug {
		subConn, err := redis.DialURL("redis://redis-service")

		if err != nil {
			log.Fatalf("Error when creating subConn: %v", err)
			return nil
		}

		log.Println("Creating subscriber connection")
		hub.subConn = &redis.PubSubConn{Conn: subConn}
		hub.subConn.Subscribe("chat")

		log.Println("Creating publisher connection")
		pubConn, err := redis.DialURL("redis://redis-service")
		if err != nil {
			log.Fatalf("Error when creating pubConn: %v", err)
			return nil
		}
		hub.pubConn = &redis.PubSubConn{Conn: pubConn}
	}

	return &hub
}

func (h *Hub) run() {
	if !Debug {
		go func() {
			for {
				switch v := h.subConn.Receive().(type) {
				case redis.Message:
					log.Printf("Message from redis: %s\n:", v.Data)
					h.broadcast <- v.Data
				case redis.Subscription:
					log.Printf("Subscription: %s: %s %d\n", v.Channel, v.Kind, v.Count)
				case error:
					log.Fatal(v)
					return
				}
			}
		}()
	}

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// todo deserialize json here
			// todo add hostname here
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
