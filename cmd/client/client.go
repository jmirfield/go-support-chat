package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/jmirfield/go-support-chat/internal/chat"
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
		client = chat.NewClient(url, n, true)
	} else {
		client = chat.NewClient(url, n, false)
	}

	client.Start()
}
