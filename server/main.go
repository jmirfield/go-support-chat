package main

import (
	"log"
	"net/http"

	"chat.server.com/internal/chat"
)

var server = chat.NewServer()

func main() {
	go server.Start()
	http.HandleFunc("/connect", server.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
