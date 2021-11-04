package main

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func (h *Hub) run() {
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

		//if err := client.WriteJSON(message); err != nil {
		//	log.Printf("error occirred: %v", err)
		//}
	}
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello world")
	})

	hub := NewHub()
	go hub.run()

	e.GET("/ws", func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			log.Println(err)
		}
		defer func() {
			delete(hub.Clients, ws)
			err := ws.Close()
			if err != nil {
				log.Printf("Closed!")
			}
			log.Printf("Closed!")
		}()
		hub.Clients[ws] = true
		log.Print("connected!")
		read(hub, ws)
		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))

}
