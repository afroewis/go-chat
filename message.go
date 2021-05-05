package main

type Message struct {
	User string `json:"user"`

	Message string `json:"message"`

	Hostname string `json:"hostname"`
}
