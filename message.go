package main

type Message struct {
	User string `json:"user"`

	Chatroom string `json:"channel"` // the chat channel in which the message was sent

	Message string `json:"message"`

	Hostname string `json:"hostname"`
}
