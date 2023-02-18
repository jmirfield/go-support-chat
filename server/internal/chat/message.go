package chat

import "fmt"

type Message struct {
	Sender string `json:"sender"`
	Body   string `json:"body"`
	ID     int    `json:"ID"`
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Body)
}

func (m Message) ToByte() []byte {
	return []byte(m.String())
}
