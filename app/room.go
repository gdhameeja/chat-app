package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var clientId = 1

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

type Room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan []byte
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
}

func NewRoom() *Room {
	return &Room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *Room) Run() {
	for {
		// works like a switch statement.
		// only one peice of block executes in one iteration
		select {
		case client := <-r.join:
			// joining
			fmt.Println("DEBUG(Run): joining", client.id)
			r.clients[client] = true
		case client := <-r.leave:
			// leaving
			fmt.Println("DEBUG(Run): leaving", client.id)
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// forward message to all clients
			fmt.Println("DEBUG(Run): received message", msg)
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

// turns type room to a http.Handler
func (r *Room) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	// upgrades the HTTP connection to web socket connection
	socket, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}

	client := &client{
		id:     clientId,
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	clientId++

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
