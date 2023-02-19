package chat

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{}

const (
	serverName = "Server"
	serverID   = -1
)

// Server represents the struct for the entire chat application
type Server struct {
	workers map[*client]*client
	queue   []*client
	message chan message
	done    chan *client
	poll    chan bool
	stop    chan bool
	mu      sync.Mutex
}

// NewServer creates manages the creation of a Server struct
func NewServer() *Server {
	return &Server{
		workers: make(map[*client]*client),
		queue:   []*client{},
		message: make(chan message, 100),
		done:    make(chan *client, 5),
		poll:    make(chan bool, 100),
		stop:    make(chan bool),
	}
}

// Start reads from the Server channels and handles incoming data accordingly
func (s *Server) Start() {
	for {
		select {
		case msg := <-s.message:
			s.write(msg)
		case client := <-s.done:
			s.unregisterClientHandler(client)
		case <-s.poll:
			s.registerNextClient()
		case <-s.stop:
			return
		}
	}
}

// Stop stops the Server
func (s *Server) Stop() {
	s.stop <- true
}

// Handler is the http handler which implements http handler func
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	c := newClient(conn, s, r.Header.Get("Name"))
	if r.Header.Get("Type") == "S" {
		s.registerSupportClient(c)
	} else {
		s.addToQueue(c)
	}
	c.start()
}

func (s *Server) send(msg message) {
	s.message <- msg
}

func (s *Server) write(msg message) {
	for k, v := range s.workers {
		if k.id == msg.id && v != nil {
			v.write(msg)
			return
		}
		if v != nil && v.id == msg.id {
			k.write(msg)
			return
		}
	}
}

func (s *Server) unregisterClientHandler(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.workers {
		if k.id == c.id {
			s.unregisterSupportClient(k)
			break
		}
		if v != nil && v.id == c.id {
			s.unregisterClient(v)
			break
		}
	}
	for _, v := range s.queue {
		if v.id == c.id {
			s.unregisterClient(v)
			break
		}
	}
}

func (s *Server) close(c *client) {
	s.done <- c
}

func (s *Server) addToQueue(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, c)
	s.poll <- true
}

func (s *Server) registerSupportClient(sc *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers[sc] = nil
	s.poll <- true
}

func (s *Server) unregisterSupportClient(sc *client) {
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return
	}

	c := s.queue[0]
	for k, v := range s.workers {
		if v == nil {
			s.workers[k] = c
			msg := newMessage(serverName, fmt.Sprintf("%s has joined the chat!", c.name), serverID)
			k.write(msg)
			s.queue = s.queue[1:]
			return
		}
	}
}

func (s *Server) unregisterClient(c *client) {
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
			msg := newMessage(serverName, fmt.Sprintf("%s has left the chat!", v.name), serverID)
			k.write(msg)
			s.poll <- true
			return
		}
	}
}
