#Default target builds the project
build:
	go build src/server.go src/client.go src/commons.go

clean:
	go clean
