package chat

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

// Client used to connect to websocket server
type Client struct {
	socket *websocket.Conn
	text   chan string
	done   chan os.Signal
}

// NewClient creates a new client with a connection to the websocket server
func NewClient(name string, support bool, url url.URL) *Client {
	headers := http.Header{}
	headers.Set("Name", name)
	if support {
		headers.Set("Type", "S")
	}

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), headers)
	if err != nil {
		log.Fatal("error connecting to server: ", err)
	}

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt, syscall.SIGINT)

	return &Client{
		socket: conn,
		text:   make(chan string),
		done:   interupt,
	}
}

// Start starts client communication with server
func (c *Client) Start() {
	defer c.Close()
	go c.read()
	go c.write()

	for {
		select {
		case <-c.done:
			return
		case text := <-c.text:
			c.send(text)
		}

	}
}

// Close closes client socket connection
func (c *Client) Close() {
	c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.socket.Close()
}

// Stop stops client execution and will eventually close the client socket connection
func (c *Client) Stop() {
	c.done <- syscall.SIGINT
}

func (c *Client) read() {
	for {
		var msg Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			// log.Println("read: ", err)
			c.Stop()
			return
		}
		log.Print(msg.String())
	}
}

func (c *Client) write() {
	for {
		var text string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text = scanner.Text()
		}
		c.text <- text
	}
}

func (c *Client) send(text string) {
	msg := Message{Body: text}
	json, _ := json.Marshal(msg)
	if err := c.socket.WriteMessage(websocket.TextMessage, json); err != nil {
		// log.Println("write: ", err)
		c.Stop()
		return
	}
}
