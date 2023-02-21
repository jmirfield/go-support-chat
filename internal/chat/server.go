package chat

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader   = &websocket.Upgrader{}
	serverUUID = uuid.New()
)

const (
	serverName = "Server"
)

// Server represents the struct for the entire chat application
type Server struct {
	workers  map[*user]*user
	queue    []*user
	messages chan Message
	close    chan uuid.UUID
	poll     chan bool
	stop     chan bool
	mu       sync.Mutex
}

// NewServer creates manages the creation of a Server struct
func NewServer() *Server {
	return &Server{
		workers:  make(map[*user]*user),
		queue:    []*user{},
		messages: make(chan Message, 5),
		close:    make(chan uuid.UUID, 5),
		poll:     make(chan bool, 5),
		stop:     make(chan bool),
	}
}

// Start reads from the Server channels and handles incoming data accordingly
func (s *Server) Start() {
	for {
		select {
		case msg := <-s.messages:
			go s.write(msg)
		case id := <-s.close:
			go s.unregisterUserHandler(id)
		case <-s.poll:
			go s.registerNextUser()
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

	name := r.Header.Get("Name")
	if r.Header.Get("Type") == "S" {
		s.newUser(conn, name, true)
	} else {
		s.newUser(conn, name, false)
	}
}

func (s *Server) newUser(conn *websocket.Conn, name string, support bool) {
	usr := &user{
		id:      uuid.New(),
		socket:  conn,
		name:    name,
		message: s.messages,
		close:   s.close,
	}

	if support {
		s.registerSupportUser(usr)
	} else {
		s.addToUserQueue(usr)
	}

	go usr.read()
}

func (s *Server) write(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.workers {
		// If support user sent the message, forward that to end user
		if k.id == msg.ID && v != nil {
			go v.write(msg)
			return
		}
		// If end user send the message, forward that to support user
		if v != nil && v.id == msg.ID {
			go k.write(msg)
			return
		}
	}
}

// Looks for the user, either in the workers map or in the queue
func (s *Server) unregisterUserHandler(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for su, u := range s.workers {
		if su.id == id {
			s.unregisterSupportUser(su)
			break
		}
		if u != nil && u.id == id {
			s.unregisterUser(u)
			break
		}
	}
	for _, u := range s.queue {
		if u.id == id {
			s.unregisterUser(u)
			break
		}
	}
}

func (s *Server) addToUserQueue(usr *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	usr.write(NewMessage(serverName, fmt.Sprint("Waiting for support user to join chat..."), serverUUID))
	s.queue = append(s.queue, usr)
	s.poll <- true
}

func (s *Server) registerSupportUser(usr *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers[usr] = nil
	s.poll <- true
}

func (s *Server) unregisterSupportUser(usr *user) {
	defer usr.socket.Close()
	u, ok := s.workers[usr]
	if !ok {
		return
	}
	if u != nil {
		u.write(NewMessage(serverName, fmt.Sprintf("%s has lost connection...", usr.name), serverUUID))
		// Bad pattern? Not sure
		s.mu.Unlock()
		s.addToUserQueue(u)
		s.mu.Lock()
	}
	delete(s.workers, usr)
}

func (s *Server) registerNextUser() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return
	}

	usr := s.queue[0]
	for su, u := range s.workers {
		if u == nil {
			s.workers[su] = usr
			su.write(NewMessage(serverName, fmt.Sprintf("%s has joined the chat!", usr.name), serverUUID))
			usr.write(NewMessage(serverName, fmt.Sprintf("%s has joined the chat!", su.name), serverUUID))
			s.queue = s.queue[1:]
			return
		}
	}
}

func (s *Server) unregisterUser(usr *user) {
	defer usr.socket.Close()
	// If user disconnects while still in queue
	for i, u := range s.queue {
		if u == usr {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			return
		}
	}
	// If user disconnects while chatting with support user
	for su, u := range s.workers {
		if u == usr {
			s.workers[su] = nil
			su.write(NewMessage(serverName, fmt.Sprintf("%s has left the chat!", usr.name), serverUUID))
			s.poll <- true
			return
		}
	}
}
