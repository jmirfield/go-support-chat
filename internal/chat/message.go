package chat

import "fmt"

type Message struct {
	Sender string `json:"sender,omitempty"`
	Body   string `json:"body"`
	ID     int    `json:"id,string,omitempty"`
}

func NewMessage(sender string, body string, id int) Message {
	return Message{Sender: sender, Body: body, ID: id}
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Body)
}

func (m Message) toByte() []byte {
	return []byte(m.String())
}
