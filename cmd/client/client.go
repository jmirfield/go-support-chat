package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/jmirfield/support-chat-websockets/internal/chat"
)

var supportFlag = flag.Bool("s", false, "true if support client")

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

	var client *chat.Client
	url := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/connect"}
	if *supportFlag {
		client = chat.NewClient(n, true, url)
	} else {
		client = chat.NewClient(n, false, url)
	}

	client.Start()
}
