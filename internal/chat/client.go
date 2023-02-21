package chat

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

// Client used to connect to websocket server
type Client struct {
	socket *websocket.Conn
	text   chan string
	done   chan os.Signal
}

// NewClient creates a new client with a connection to the websocket server
func NewClient(url url.URL, name string, support bool) *Client {
	headers := http.Header{}
	headers.Set("Name", name)
	if support {
		headers.Set("Type", "S")
	}

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), headers)
	if err != nil {
		log.Fatal("error connecting to server: ", err)
	}

	return &Client{
		socket: conn,
		text:   make(chan string),
	}
}

// Start starts client communication with server
func (c *Client) Start() {
	defer c.close()
	go c.read()

	var scanner = bufio.NewScanner(os.Stdin)
	var text string
	for {
		if scanner.Scan() {
			text = scanner.Text()
			c.write(text)
		}
	}
}

// Close closes client socket connection
func (c *Client) close() {
	c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.socket.Close()
	os.Exit(-1)
}

func (c *Client) read() {
	for {
		var msg Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			// log.Println("read: ", err)
			c.close()
			return
		}
		log.Print(msg)
	}
}

func (c *Client) write(text string) {
	msg := Message{Body: text}
	if err := c.socket.WriteJSON(msg); err != nil {
		// log.Println("write: ", err)
		c.close()
		return
	}
}
