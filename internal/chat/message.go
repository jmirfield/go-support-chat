package chat

import "fmt"

// Message holds message information
type Message struct {
	Sender string `json:"sender,omitempty"`
	Body   string `json:"body"`
	ID     int    `json:"id,string,omitempty"`
}

// NewMessage creates a Message
func NewMessage(sender string, body string, id int) Message {
	return Message{Sender: sender, Body: body, ID: id}
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Body)
}
