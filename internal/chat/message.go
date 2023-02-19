package chat

import "fmt"

type message struct {
	sender string
	body   string
	id     int
}

func newMessage(sender string, body string, id int) message {
	return message{sender, body, id}
}

func (m message) String() string {
	return fmt.Sprintf("%s: %s", m.sender, m.body)
}

func (m message) toByte() []byte {
	return []byte(m.String())
}
