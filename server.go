package main

import (
	"fmt"
	"log"
	"net"

	db "cse312.app/database"
	httpServer "cse312.app/http"
)

var HOST = "mongo"

//https://golang.org/pkg/net/#Listener
func main() {
	l, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	//init database connection
	if err := db.StartDB(HOST); err != nil {
		panic(err)
	}
	fmt.Println("Server Started")

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection Accepted")
		go httpServer.HandleConnection(conn)
	}
}
