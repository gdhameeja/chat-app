package app

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user
type client struct {
	id int
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	room *Room
}

func (c *client) read() {
	fmt.Printf("DEBUG: client %d read gets called\n", c.id)
	defer c.socket.Close()
	for {
		fmt.Println("DEBUG(client): Read loop")
		_, msg, err := c.socket.ReadMessage()
		fmt.Println("DEBUG(client): incoming message to client", c.id, msg)
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	fmt.Printf("DEBUG: client %d write gets called\n", c.id)
	defer c.socket.Close()
	for msg := range c.send {
		fmt.Println("DEBUG(client): client is writing messages ", c.id, msg)
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
