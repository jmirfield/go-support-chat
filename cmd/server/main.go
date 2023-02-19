package main

import (
	"log"
	"net/http"

	"github.com/jmirfield/support-chat-websockets/internal/chat"
)

func main() {
	server := chat.NewServer()
	go server.Start()
	http.HandleFunc("/connect", server.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
