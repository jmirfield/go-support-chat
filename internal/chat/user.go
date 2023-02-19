package chat

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var mu sync.Mutex

type user struct {
	socket  *websocket.Conn
	server  *Server
	name    string
	id      int
	message chan Message
	done    chan bool
}

var id = 0

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

func (c *user) close() {
	c.done <- true
}

func (c *user) write(msg Message) {
	c.message <- msg
}

func (c *user) writeTo() {
	for {
		select {
		case msg := <-c.message:
			json, _ := json.Marshal(msg)
			if err := c.socket.WriteMessage(websocket.TextMessage, json); err != nil {
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

func (c *user) readFrom() {
	for {
		var msg Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			log.Println("client read: ", err)
			c.close()
			return
		}
		msg = NewMessage(c.name, string(msg.Body), c.id)
		c.server.send(msg)
	}
}
