package main

import (
	"fmt"
	"math/rand"
	"net"
)

func main() {
	port := rand.Intn(65535-1025) + 1025
	ln, _ := net.Listen("udp", fmt.Printf(":%d", port))

	for {
		conn, _ := ln.Accept()

		go handleConnection(conn)
	}

}

func handleConnection(conn Conn) {

}
