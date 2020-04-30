The DockerHost server to handle the docker stuff

# The mechanics
## Basic steps in handling every connection: 
1. Connection
2. Request (to a client)
3. Feedback (to the server)
## The main data stream diagram
1. Client connected to the server
2. Server sends a message to a client with RPC Request 
3. Client handles the Request and sends a Feedback
4. Server manages the client's Feedback

# Preparation and Installation
`go get github.com/gorilla`

`go get github.com/docker/docker/client`

`go get github.com/satori/go.uuid`

`go get github.com/mitchellh/mapstructure`


# Testing
`go test -v`
