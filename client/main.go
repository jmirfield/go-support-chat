package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var sFlag = flag.Bool("s", false, "true if support client")

func main() {
	flag.Parse()
	var n string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "What is your name? ")
		n, _ = r.ReadString('\n')
		if n != "" {
			break
		}
	}

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt)

	url := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/connect"}
	headers := http.Header{}
	headers.Set("Name", n)
	if *sFlag {
		headers.Set("Type", "S")
	}

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), headers)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read: ", err)
				break
			}
			log.Printf("%s", message)
		}
	}()

	input := make(chan string)

	go func() {
		defer close(input)
		for {
			var t string
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				t = scanner.Text()
			}
			if err != nil {
				log.Println(err)
				break
			}
			input <- t
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interupt:
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		case i := <-input:
			err := conn.WriteMessage(websocket.TextMessage, []byte(i))
			if err != nil {
				log.Println("write: ", err)
				return
			}
		}

	}
}
