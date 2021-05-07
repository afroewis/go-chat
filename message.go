package main

type Message struct {
	User string `json:"username"`

	Chatroom string `json:"chatroom"`

	Message string `json:"message"`

	Hostname string `json:"hostname"`
}
