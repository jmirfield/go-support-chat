package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

var mu sync.Mutex

type client struct {
	socket  *websocket.Conn
	server  *Server
	name    string
	id      int
	message chan message
	done    chan bool
}

var id = 0

func newClient(conn *websocket.Conn, server *Server, name string) *client {
	mu.Lock()
	defer mu.Unlock()
	id++
	return &client{
		socket:  conn,
		server:  server,
		name:    name,
		id:      id,
		message: make(chan message),
		done:    make(chan bool),
	}
}

func (c *client) start() {
	go c.writeTo()
	go c.readFrom()
}

func (c *client) close() {
	c.done <- true
}

func (c *client) write(msg message) {
	c.message <- msg
}

func (c *client) writeTo() {
	for {
		select {
		case msg := <-c.message:
			if err := c.socket.WriteMessage(websocket.TextMessage, msg.toByte()); err != nil {
				// log.Println("client write: ", err)
				c.close()
				return
			}
		case <-c.done:
			c.server.close(c)
			return
		}
	}
}

func (c *client) readFrom() {
	for {
		_, m, err := c.socket.ReadMessage()
		if err != nil {
			// log.Println("client read: ", err)
			c.close()
			return
		}
		msg := newMessage(c.name, string(m), c.id)
		c.server.send(msg)
	}
}
