package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

var WINDOW_SIZE uint16 = 10

const PACKET_SIZE = 1000

var ACK_NUMBER uint32 = 0

var dataChunks [][]byte

var inflight = make(map[int]time.Time)

//var TEN = 10
//var ONE_HUNDRED = 100
//var ONE_THOUSAND = TEN * ONE_HUNDRED
//var ONE_MILLION = ONE_THOUSAND * ONE_THOUSAND

var retries chan uint32
var unsent chan uint32

var conn net.Conn

func main() {
	hostPort := os.Args[1]
	splitList := strings.Split(hostPort, ":")
	host := splitList[0]
	port := splitList[1]
	_, _ = host, port

	conn, _ = net.Dial("udp", hostPort)
	data, _ := ioutil.ReadAll(os.Stdin)

	retries = make(chan uint32, (len(data)/PACKET_SIZE)+1)
	unsent = make(chan uint32, (len(data)/PACKET_SIZE)+1)

	for i := 0; i < len(data); i++ {
		start := i * PACKET_SIZE
		end := min(len(data), start+PACKET_SIZE)
		dataChunks = append(dataChunks, data[start:end])
		unsent <- uint32(i)
	}

	// goroutine for acks -> update table of in flight
	go updateAcks()

	// routine to do sending
	go sendDataChunks()

	// routine to check for timeouts and queue resends
	go checkForTimeouts()

	for {
		// gg
	}
}

func updateAcks() {

}

func sendDataChunks() {
	for {
		if len(retries) > 0 {
			data := <-retries
			sendData(data)
		} else if len(unsent) > 0 {
			data := <-unsent
			sendData(data)
		}
	}
}

func sendData(data uint32) {
	packet := Packet{
		Seq:       data,
		Ack:       ACK_NUMBER,
		AdvWindow: WINDOW_SIZE,
		Flags:     0, // TODO
		Data:      dataChunks[data],
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet)
	fmt.Fprint(conn, buf.Bytes())
}

func checkForTimeouts() {

}
