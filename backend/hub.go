package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type Hub struct {
	Clients   map[*websocket.Conn]bool
	Broadcast chan Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if err := client.WriteJSON(message); err != nil {
					log.Printf("error occurred: %v", err)
				}
			}
		}
	}
}

func read(hub *Hub, client *websocket.Conn) {
	for {
		var message Message
		err := client.ReadJSON(&message)
		if err != nil {
			log.Printf("error occurred: %v", err)
			delete(hub.Clients, client)
			break
		}
		log.Println(message)
		hub.Broadcast <- message
	}
}
