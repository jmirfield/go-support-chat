package main

import (
	"log"
	"net/http"

	"github.com/jmirfield/go-support-chat/internal/chat"
)

func main() {
	server := chat.NewServer()
	go server.Start()
	http.HandleFunc("/connect", server.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
