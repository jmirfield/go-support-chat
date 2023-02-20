package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

type user struct {
	socket  *websocket.Conn
	server  *Server
	name    string
	id      int
	message chan Message
	done    chan bool
}

var (
	mu sync.Mutex
	id = 0
)

func newUser(conn *websocket.Conn, server *Server, name string) *user {
	mu.Lock()
	defer mu.Unlock()
	id++
	return &user{
		socket:  conn,
		server:  server,
		name:    name,
		id:      id,
		message: make(chan Message, 5),
		done:    make(chan bool),
	}
}

func (c *user) start() {
	go c.writeTo()
	go c.readFrom()
}

func (c *user) writeTo() {
	for {
		select {
		case msg := <-c.message:
			if err := c.socket.WriteJSON(msg); err != nil {
				// log.Println("client write: ", err)
				c.done <- true
				return
			}
		case <-c.done:
			c.server.done <- c
			return
		}
	}
}

func (c *user) readFrom() {
	for {
		var msg Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			// log.Println("client read: ", err)
			c.done <- true
			return
		}
		msg = NewMessage(c.name, string(msg.Body), c.id)
		c.server.message <- msg
	}
}
