package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Message string `json:"message"`
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "sample message",
		})
	})

	hub := NewHub()
	go hub.Run()

	r.GET("/ws", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
		}
		defer func() {
			delete(hub.Clients, ws)
			err := ws.Close()
			if err != nil {
				log.Printf("Closed!")
			}
		}()

		hub.Clients[ws] = true
		log.Print("connected!")
		read(hub, ws)
	})
	err := r.Run()
	if err != nil {
		return
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
