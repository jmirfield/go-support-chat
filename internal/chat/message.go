package chat

import (
	"fmt"

	"github.com/google/uuid"
)

// Message holds message information
type Message struct {
	Sender string    `json:"sender,omitempty"`
	Body   string    `json:"body"`
	ID     uuid.UUID `json:"id,string,omitempty"`
}

// NewMessage creates a Message
func NewMessage(sender string, body string, id uuid.UUID) Message {
	return Message{Sender: sender, Body: body, ID: id}
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Body)
}
