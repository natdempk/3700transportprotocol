package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"

	"../tpl"
)

// how do we want to handle non contiguous packets
// and out of order delivery?
var dataChunks = make(map[uint32][]byte)

var done = false

var WINDOW_SIZE uint16 = 10
var conn net.PacketConn

func main() {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(65535-1025) + 1025
	conn, _ = net.ListenPacket("udp", fmt.Sprintf(":%d", port))

	tpl.Log("[bound] %d\n", port)

	// We need to set up a listener socket
	// And also a sender socket

	for !done {
		packet, retAddr := tpl.ReadPacket(conn)
		go handleConnection(packet, retAddr)
	}

	// we're done

	for i := 0; i < len(dataChunks); i++ {
		fmt.Printf("%s", dataChunks[uint32(i)])
	}

	tpl.Log("[completed]")
}

func haveAllPackets(seq uint32) bool {
	return len(dataChunks) == int(seq)+1
}

func getStatus(seq uint32) string {
	return "ACCEPTED (in-order)"
}

func handleConnection(packet tpl.Packet, retAddr net.Addr) {
	// store data in a map
	dataChunks[packet.Seq] = packet.Data[:packet.Size]

	tpl.Log("[recv data] %v (%v) %v", packet.Seq*tpl.PACKET_SIZE, len(packet.Data), getStatus(packet.Seq))
	var flag uint16 = 2
	if packet.Flags == 1 && haveAllPackets(packet.Seq) {
		done = true
		flag = 3
		// TODO: add a final shutdown flag thing
	}

	var data [1024]byte
	// send an acknowledgement packet
	acket := tpl.Packet{
		Seq:       packet.Seq,
		Size:      0,
		Ack:       packet.Seq,
		AdvWindow: WINDOW_SIZE,
		Flags:     flag,
		Data:      data,
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &acket)

	conn.WriteTo(buf.Bytes(), retAddr)

	// the issue with the last packet is it could be dropped during delivery

	// so we might actually need a 3 way handshake or something around closing out
}
