# go-support-chat
This application was for golang practice with websockets. 

My main goal was to create a live-chat support application where end users can be paired up with support users for help. Support users will wait for end users to join and if there are no support users available, end users wait in a queue.

Looking for feedback!

# Design
Server in the internal/chat package does the bulk of the work. All messages get sent to the Server and flow to the appropriate client(s). Client in internal/chat package is used for both support and end user clients. If a support user registers, Server moves them into the workers map waiting for an end user to join. If an end user joins, they get sent to the user queue. The queue is polled on certain actions like if a support user registers, or if an end user disconnects and if there is an available support user, the end user gets removed from the queue and connected.

# Usage
You will need to run at least 3 terminals to play around with this.

To run the server:
```
go run ./cmd/server
```

To run client as an end user:
```
go run ./cmd/client
```

To run client as a support user:
```
go run ./cmd/client -s
```
