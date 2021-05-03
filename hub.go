package main

import (
	"fmt"
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
	subConn, err := redis.DialURL("redis://redis-service")

	if err != nil {
		log.Fatal("Error when creating subConn: %v", err)
		return nil
	}

	subscriberPsc := redis.PubSubConn{Conn: subConn}
	log.Println("Created subscriber connection")
	subscriberPsc.Subscribe("chat")

	pubConn, err := redis.DialURL("redis://redis-service")
	log.Println("Created publisher connection")
	if err != nil {
		log.Fatal("Error when creating pubConn: %v", err)
		return nil
	}
	publisherPsc := redis.PubSubConn{Conn: pubConn}

	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		subConn:    &subscriberPsc,
		pubConn:    &publisherPsc,
	}
}

func (h *Hub) run() {
	go func() {
		for {
			switch v := h.subConn.Receive().(type) {
			case redis.Message:
				fmt.Printf("Message from redis: %s\n:", v.Data)
				h.broadcast <- v.Data
			case redis.Subscription:
				fmt.Printf("Subscription: %s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				log.Fatal(v)
				return
			}
		}
	}()

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
