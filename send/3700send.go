package main

import (
	"io/ioutil"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"../tpl"
)

var WINDOW_SIZE uint32 = 24

var done = false

var timeOut = 1 * time.Second
var ACK_NUMBER uint32 = 0

var dataChunks [][]byte
var dataSizes = make(map[uint32]uint32)

var inflight = make(map[uint32]time.Time)
var inflightMutex = &sync.Mutex{}

var retries chan uint32
var unsent chan uint32

var conn net.Conn
var rtt time.Duration = timeOut
var c = 0.4

var binv float64 = 2

var cwndMax int = int(WINDOW_SIZE)

var lastWindowRed = time.Now()
var recvOrSentPacket = time.Now()

func getCwnd() (cwnd int) {
	t := float64(time.Since(lastWindowRed).Seconds())

	cube := math.Cbrt(float64(cwndMax / 2))
	cwnd = int(math.Floor(math.Pow(3, t-cube)*c + float64(cwndMax)))
	return
}

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
	defer conn.Close()

	data, _ := ioutil.ReadAll(os.Stdin)

	retries = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)
	unsent = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)

	for i := 0; i < len(data)/tpl.PACKET_SIZE+1; i++ {
		start := i * tpl.PACKET_SIZE
		end := tpl.Min(len(data), start+tpl.PACKET_SIZE)
		dataChunks = append(dataChunks, data[start:end])
		dataSizes[uint32(i)] = uint32(end - start)
		unsent <- uint32(i)
	}

	// goroutine for acks -> update table of in flight
	go updateAcks()

	// routine to do sending
	go sendDataChunks()

	// routine to check for timeouts and queue resends
	go checkForTimeouts()

	for !done || time.Since(recvOrSentPacket).Seconds() > 5 {
		// keep going
		time.Sleep(10 * time.Millisecond)
	}

	tpl.Log("finished")
	var emptyData []byte
	packet := tpl.Packet{
		Seq:       1,
		Size:      0,
		Ack:       0,
		AdvWindow: WINDOW_SIZE,
		Flags:     4,
		Data:      emptyData,
	}

	buf := tpl.WriteBytes(packet)
	conn.Write(buf.Bytes())

	tpl.Log("[completed]")
}

func updateAcks() {
	for !done {
		packet := tpl.ReadPacketC(conn)
		recvOrSentPacket = time.Now()
		if val, ok := inflight[packet.Ack]; ok {

			alpha := 0.875

			rtt = time.Duration(alpha*float64(rtt.Nanoseconds()) + (1-alpha)*float64(time.Since(val).Nanoseconds()))
			deleteInflight(packet.Ack)
		}
		tpl.Log("[recv ack] %v", packet.Ack*tpl.PACKET_SIZE)
		// TODO: optimizations
		done = done || packet.Flags == 3
	}
	tpl.Log("done with acks")
}

func sendDataChunks() {
	for !done {

		if len(inflight) < getCwnd() {
			if len(retries) > 0 {
				data := <-retries
				sendData(data)
			} else if len(unsent) > 0 {
				data := <-unsent
				sendData(data)
			}
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func sendData(data uint32) {

	recvOrSentPacket = time.Now()
	var flags uint32 = 0
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
	buf := tpl.WriteBytes(packet)
	if !done {
		conn.Write(buf.Bytes())
		tpl.Log("[send data] %v (%v)", packet.Seq*tpl.PACKET_SIZE, len(packet.Data))
	}
}

func checkForTimeouts() {
	for !done {
		for seq, sendTime := range inflight {
			if time.Since(sendTime) >= 2*rtt {
				cwndMax = getCwnd()
				lastWindowRed = time.Now()
				retries <- seq
				WINDOW_SIZE = 1
				deleteInflight(seq)
			}
		}

		time.Sleep(1000 * time.Millisecond) // will we change this number? will it matter?
	}
}
