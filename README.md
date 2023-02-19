# support-chat-golang
This application was for golang practice with websockets. My main goal was to create a live support application where support users can sit and wait for end users to join the chat. End users wait in a queue if there are not any support users available.

Looking for feedback!

# Usage
You will need to run at least 3 terminals to play around with this.

To run the server:
```
go run ./cmd/server
```

To run a client:
```
// for end user
go run ./cmd/client

// for support user
go run ./cmd/client -s
```