package main

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"../tpl"
)

var WINDOW_SIZE uint16 = 10

var done = false

var timeOut = 100 * time.Millisecond
var ACK_NUMBER uint32 = 0

var dataChunks [][tpl.PACKET_SIZE]byte
var dataSizes = make(map[uint32]uint16)

var inflight = make(map[uint32]time.Time)
var inflightMutex = &sync.Mutex{}

var retries chan uint32
var unsent chan uint32

var conn net.Conn

func setInflight(i uint32) {
	inflightMutex.Lock()
	inflight[i] = time.Now()
	inflightMutex.Unlock()
	return
}

func deleteInflight(i uint32) {
	inflightMutex.Lock()
	delete(inflight, i)
	inflightMutex.Unlock()
	return
}

func main() {
	hostPort := os.Args[1]

	conn, _ = net.Dial("udp", hostPort)

	data, _ := ioutil.ReadAll(os.Stdin)

	retries = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)
	unsent = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)

	for i := 0; i < len(data)/tpl.PACKET_SIZE+1; i++ {
		start := i * tpl.PACKET_SIZE
		var s [tpl.PACKET_SIZE]byte
		end := tpl.Min(len(data), start+tpl.PACKET_SIZE)
		for start := i * tpl.PACKET_SIZE; start < end; start++ {
			s[start%tpl.PACKET_SIZE] = data[start]
		}
		dataChunks = append(dataChunks, s)
		dataSizes[uint32(i)] = uint16(end - start)
		unsent <- uint32(i)
	}

	// goroutine for acks -> update table of in flight
	go updateAcks()

	// routine to do sending
	go sendDataChunks()

	// routine to check for timeouts and queue resends
	go checkForTimeouts()

	for !done {
		// keep going
		time.Sleep(10 * time.Millisecond)
	}

	tpl.Log("[completed]")
}

func updateAcks() {
	for !done {
		packet := tpl.ReadPacketC(conn)
		deleteInflight(packet.Ack)
		tpl.Log("[recv ack] %v", packet.Ack*tpl.PACKET_SIZE)
		// TODO: optimizations
		done = packet.Flags == 3
	}
}

func sendDataChunks() {
	for !done {
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
	var flags uint16 = 0
	if data == uint32(len(dataChunks)-1) {
		flags = 1 // we're done
	}
	size := dataSizes[data]
	packet := tpl.Packet{
		Seq:       data,
		Size:      size,
		Ack:       ACK_NUMBER,
		AdvWindow: WINDOW_SIZE,
		Flags:     flags,
		Data:      dataChunks[data],
	}

	setInflight(data)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &packet)
	if !done {
		conn.Write(buf.Bytes())
		tpl.Log("[send data] %v (%v)", packet.Seq*tpl.PACKET_SIZE, len(packet.Data))
	}
}

func checkForTimeouts() {
	for !done {
		for seq, sendTime := range inflight {
			if time.Since(sendTime)*time.Millisecond >= timeOut {
				retries <- seq
				deleteInflight(seq)
			}
		}

		time.Sleep(1000 * time.Millisecond) // will we change this number? will it matter?
	}
}
