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

func main() {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(65535-1025) + 1025
	udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
	}

	tpl.Log("[bound] %d\n", port)

	//_ = ln

	for !done {
		fmt.Println("aa")
		//conn, err := ln.Accept()
		//if err != nil {
		//fmt.Println(err)
		//}

		go handleConnection(conn)
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

func handleConnection(conn net.Conn) {
	packet := tpl.ReadPacket(conn)
	// store data in a map
	dataChunks[packet.Seq] = packet.Data

	tpl.Log("[recv data] %v (%v) %v", packet.Seq*tpl.PACKET_SIZE, len(packet.Data), getStatus(packet.Seq))

	if packet.Flags == 1 && haveAllPackets(packet.Seq) {
		done = true
		// TODO: add a final shutdown flag thing
	}

	var data []byte
	// send an acknowledgement packet
	acket := tpl.Packet{
		Seq:       packet.Seq,
		Ack:       packet.Seq,
		AdvWindow: WINDOW_SIZE,
		Flags:     2,
		Data:      data,
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, acket)

	fmt.Fprint(conn, buf.Bytes())

	// the issue with the last packet is it could be dropped during delivery

	// so we might actually need a 3 way handshake or something around closing out
}
