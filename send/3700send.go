package main

import (
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"../tpl"
)

var done = false
var dataChunks [][]byte
var retries chan uint32
var unsent chan uint32
var conn net.Conn

var initRtt = 100 * time.Millisecond
var rtt = initRtt

// Used to manage a final timeout and signal total failure
var recvOrSentPacket = time.Now()

var c = 0.4
var binv float64 = 2
var cwndMax = 15

var lastWindowRed = time.Now()

// consider moving this to another file
func getCwnd() (cwnd int) {
	t := float64(time.Since(lastWindowRed).Seconds())

	cube := math.Cbrt(float64(cwndMax / 2))
	cwnd = int(math.Floor(math.Pow(3, t-cube)*c + float64(cwndMax)))
	return
}

var inflight = make(map[uint32]time.Time)
var inflightMutex = &sync.Mutex{}

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

func getMinPendingAck() (seq uint32) {
	inflightMutex.Lock()
	seq = math.MaxUint32

	for i, _ := range inflight {
		seq = uint32(tpl.Min(int(seq), int(i)))
	}

	inflightMutex.Unlock()
	return
}

var ignoreLast bool

func main() {
	hostPort := os.Args[1]

	conn, _ = net.Dial("udp", hostPort)
	defer conn.Close()

	data, _ := ioutil.ReadAll(os.Stdin)
	data, ignoreLast = tpl.CompressBytes(data)

	retries = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)
	unsent = make(chan uint32, (len(data)/tpl.PACKET_SIZE)+1)

	for i := 0; i < len(data)/tpl.PACKET_SIZE+1; i++ {
		start := i * tpl.PACKET_SIZE
		end := tpl.Min(len(data), start+tpl.PACKET_SIZE)
		dataChunks = append(dataChunks, data[start:end])
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

	var emptyData []byte
	packet := tpl.Packet{
		Seq:   1,
		Flags: 4,
		Data:  emptyData,
	}

	buf := tpl.WriteBytes(packet)
	conn.Write(buf.Bytes())
	// fingers crossed
	conn.Write(buf.Bytes())

	tpl.Log("[completed]")
}

func updateAcks() {
	// keep track of min pending Ack
	var minPendingAck uint32 = math.MaxUint32

	for !done {
		packet, err := tpl.ReadPacketC(conn)
		if err != nil && err != io.EOF { // other side tore down connection, we're done
			done = true
			break
		}
		recvOrSentPacket = time.Now()

		curMinPendingAck := getMinPendingAck()
		if minPendingAck != math.MaxUint32 && minPendingAck == curMinPendingAck {
			// we have seen two acks without the min pending ack changing, so fast retransmit
			retries <- curMinPendingAck
		}

		minPendingAck = curMinPendingAck

		if val, ok := inflight[packet.Seq]; ok {
			alpha := 0.875

			if rtt == initRtt {
				rtt = time.Since(val)
			} else {
				rtt = time.Duration(alpha*float64(rtt.Nanoseconds()) + (1-alpha)*float64(time.Since(val).Nanoseconds()))
			}

			deleteInflight(packet.Seq)
		}
		tpl.Log("[recv ack] %v %v", packet.Seq*tpl.PACKET_SIZE, packet.Flags)
		done = done || packet.Flags == 3
	}
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
	var flags uint8 = 0
	if data == uint32(len(dataChunks)-1) {
		if ignoreLast {
			flags = 5
		} else {
			flags = 1 // Last data Packet
		}
	}
	packet := tpl.Packet{
		Seq:   data,
		Flags: flags,
		Data:  dataChunks[data],
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
				deleteInflight(seq)
			}
		}

		time.Sleep(1000 * time.Millisecond)
	}
}
