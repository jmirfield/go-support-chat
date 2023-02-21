package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type user struct {
	id      uuid.UUID
	socket  *websocket.Conn
	name    string
	message chan<- Message
	close   chan<- uuid.UUID
}

func (c *user) read() {
	for {
		var msg Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			// log.Println("client read: ", err)
			c.close <- c.id
			return
		}
		c.message <- NewMessage(c.name, string(msg.Body), c.id)
	}
}

func (c *user) write(msg Message) {
	if err := c.socket.WriteJSON(msg); err != nil {
		// log.Println("client write: ", err)
		c.close <- c.id
		return
	}
}
