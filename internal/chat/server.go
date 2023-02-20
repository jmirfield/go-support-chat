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
	workers map[*user]*user
	queue   []*user
	message chan Message
	done    chan *user
	poll    chan bool
	stop    chan bool
	mu      sync.Mutex
}

// NewServer creates manages the creation of a Server struct
func NewServer() *Server {
	return &Server{
		workers: make(map[*user]*user),
		queue:   []*user{},
		message: make(chan Message, 100),
		done:    make(chan *user, 5),
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
		case user := <-s.done:
			s.unregisterUserHandler(user)
		case <-s.poll:
			s.registerNextUser()
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
	u := newUser(conn, s, r.Header.Get("Name"))
	if r.Header.Get("Type") == "S" {
		s.registerSupportUser(u)
	} else {
		s.addToQueue(u)
	}
	u.start()
}

func (s *Server) write(msg Message) {
	for k, v := range s.workers {
		// If support user sent the message, forward that to end user
		if k.id == msg.ID && v != nil {
			v.message <- msg
			return
		}
		// If end user send the message, forward that to support user
		if v != nil && v.id == msg.ID {
			k.message <- msg
			return
		}
	}
}

// Looks for the user, either in the workers map or in the queue
func (s *Server) unregisterUserHandler(u *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.workers {
		if k.id == u.id {
			s.unregisterSupportUser(k)
			break
		}
		if v != nil && v.id == u.id {
			s.unregisterUser(v)
			break
		}
	}
	for _, v := range s.queue {
		if v.id == u.id {
			s.unregisterUser(v)
			break
		}
	}
}

func (s *Server) addToQueue(u *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := NewMessage(serverName, fmt.Sprint("Waiting for support user to join chat..."), serverID)
	u.message <- msg
	s.queue = append(s.queue, u)
	s.poll <- true
}

func (s *Server) registerSupportUser(su *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers[su] = nil
	s.poll <- true
}

func (s *Server) unregisterSupportUser(su *user) {
	defer su.socket.Close()
	user, ok := s.workers[su]
	if !ok {
		return
	}
	if user != nil {
		msg := NewMessage(serverName, fmt.Sprintf("%s has lost connection...", su.name), serverID)
		user.message <- msg
		// Bad pattern? Not sure
		s.mu.Unlock()
		s.addToQueue(user)
		s.mu.Lock()
	}
	delete(s.workers, su)
}

func (s *Server) registerNextUser() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return
	}

	u := s.queue[0]
	for k, v := range s.workers {
		if v == nil {
			s.workers[k] = u
			msg1 := NewMessage(serverName, fmt.Sprintf("%s has joined the chat!", u.name), serverID)
			k.message <- msg1
			msg2 := NewMessage(serverName, fmt.Sprintf("%s has joined the chat!", k.name), serverID)
			u.message <- msg2
			s.queue = s.queue[1:]
			return
		}
	}
}

func (s *Server) unregisterUser(u *user) {
	defer u.socket.Close()
	// If user disconnects while still in queue
	for i, v := range s.queue {
		if v == u {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			return
		}
	}
	// If user disconnects while chatting with support user
	for k, v := range s.workers {
		if v == u {
			s.workers[k] = nil
			msg := NewMessage(serverName, fmt.Sprintf("%s has left the chat!", v.name), serverID)
			k.message <- msg
			s.poll <- true
			return
		}
	}
}
