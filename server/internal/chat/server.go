package chat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{}

const server = "Server"

type Server struct {
	workers map[*Client]*Client
	queue   []*Client
	message chan Message
	done    chan Message
	poll    chan bool
}

func NewServer() *Server {
	return &Server{
		workers: make(map[*Client]*Client),
		queue:   []*Client{},
		message: make(chan Message),
		done:    make(chan Message),
		poll:    make(chan bool),
	}
}

func (s *Server) Start() {
	for {
		select {
		case msg := <-s.message:
			s.send(msg)
		case msg := <-s.done:
			for k, v := range s.workers {
				if k.id == msg.ID {
					s.unregisterSupportClient(k)
					break
				}
				if v != nil && v.id == msg.ID {
					s.unregisterClient(v)
					break
				}
			}
			for _, v := range s.queue {
				if v.id == msg.ID {
					s.unregisterClient(v)
					break
				}
			}
		case <-s.poll:
			s.registerNextClient()
		}
	}
}

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	c := NewClient(conn, s, r.Header.Get("Name"))
	if r.Header.Get("Type") == "S" {
		s.RegisterSupportClient(c)
	} else {
		s.AddToQueue(c)
	}
	go c.Start()
}

func (s *Server) send(msg Message) {
	for k, v := range s.workers {
		if k.id == msg.ID && v != nil {
			v.Write(msg)
			return
		}
		if v != nil && v.id == msg.ID {
			k.Write(msg)
			return
		}
	}
}

func (s *Server) AddToQueue(c *Client) {
	s.queue = append(s.queue, c)
	s.poll <- true
}

func (s *Server) RegisterSupportClient(sc *Client) {
	s.workers[sc] = nil
	s.poll <- true
}

func (s *Server) unregisterSupportClient(sc *Client) {
	defer sc.socket.Close()
	client, ok := s.workers[sc]
	if !ok {
		return
	}
	if client != nil {
		client.socket.Close()
	}
	delete(s.workers, sc)
}

func (s *Server) registerNextClient() {
	if len(s.queue) == 0 {
		return
	}

	c := s.queue[0]
	for k, v := range s.workers {
		if v == nil {
			s.workers[k] = c
			body := fmt.Sprintf("%s has joined the chat!", c.name)
			k.Write(Message{Sender: server, Body: body})
			s.queue = s.queue[1:]
			return
		}
	}
}

func (s *Server) unregisterClient(c *Client) {
	defer c.socket.Close()
	// If client disconnects while still in queue
	for i, v := range s.queue {
		if v == c {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			return
		}
	}
	// If client disconnects while chatting with support client
	for k, v := range s.workers {
		if v == c {
			s.workers[k] = nil
			body := fmt.Sprintf("%s has left the chat!", v.name)
			k.Write(Message{Sender: server, Body: body})
			go s.checkQueue()
			return
		}
	}
}

func (s *Server) checkQueue() {
	s.poll <- true
}
