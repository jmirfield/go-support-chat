package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	socket  *websocket.Conn
	server  *Server
	name    string
	id      int
	message chan Message
	done    chan bool
}

var id = 0

func NewClient(conn *websocket.Conn, server *Server, name string) *Client {
	id++
	return &Client{
		socket:  conn,
		server:  server,
		name:    name,
		id:      id,
		message: make(chan Message),
		done:    make(chan bool),
	}
}

func (c *Client) Conn() *websocket.Conn {
	return c.socket
}

func (c *Client) Start() {
	go c.writeTo()
	go c.readFrom()
}

func (c *Client) Write(msg Message) {
	c.message <- msg
}

func (c *Client) writeTo() {
	for {
		select {
		case msg := <-c.message:
			if err := c.socket.WriteMessage(websocket.TextMessage, msg.ToByte()); err != nil {
				log.Println("client write: ", err)
				c.done <- true
				return
			}
		case <-c.done:
			c.server.done <- Message{ID: c.id}
			return
		}
	}
}

func (c *Client) readFrom() {
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			log.Println("client read: ", err)
			c.done <- true
			return
		}
		c.server.message <- Message{Sender: c.name, ID: c.id, Body: string(msg)}
	}
}
