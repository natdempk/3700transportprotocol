package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func timestamp() (timestamp string) {
	nanos := time.Now().UnixNano()
	micros := nanos / 1000
	timestamp = string(micros)
	return
}

func main() {
	port := rand.Intn(65535-1025) + 1025
	ln, _ := net.Listen("udp", fmt.Sprintf(":%d", port))

	fmt.Println(time.Now())

	for {
		conn, _ := ln.Accept()

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	// TODO
}
