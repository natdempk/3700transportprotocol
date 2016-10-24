package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"../tpl"
)

// how do we want to handle non contiguous packets
// and out of order delivery?
var dataChunks = make(map[uint32][]byte)

var done = false

var finalPacketId = -1
var WINDOW_SIZE uint16 = 10
var conn net.PacketConn

func main() {
	rand.Seed(time.Now().UnixNano())
	conn, _ = net.ListenUDP("udp", nil)
	port, _ := strconv.Atoi(conn.LocalAddr().String()[5:])
	defer conn.Close()
	tpl.Log("[bound] %d\n", port)

	// We need to set up a listener socket
	// And also a sender socket

	for !done {
		packet, retAddr := tpl.ReadPacket(conn)
		handleConnection(packet, retAddr)
	}

	// we're done

	for i := 0; i < len(dataChunks); i++ {
		fmt.Printf("%s", dataChunks[uint32(i)])
	}
	var ackDone bool = false
	for !ackDone {
		packet, retAddr := tpl.ReadPacket(conn)
		if packet.Flags != 4 {

			var data [tpl.PACKET_SIZE]byte
			teardown := tpl.Packet{
				Seq:       packet.Seq,
				Size:      0,
				Ack:       packet.Seq,
				AdvWindow: WINDOW_SIZE,
				Flags:     3,
				Data:      data,
			}

			buf := new(bytes.Buffer)
			binary.Write(buf, binary.LittleEndian, &teardown)

			conn.WriteTo(buf.Bytes(), retAddr)

		} else {
			ackDone = true
		}
	}

	tpl.Log("[completed]")
	return
}

func haveAllPackets(seq int) bool {
	return len(dataChunks) == seq+1
}

func getStatus(seq uint32) string {
	return "ACCEPTED (in-order)"
}

func handleConnection(packet tpl.Packet, retAddr net.Addr) {
	// store data in a map
	dataChunks[packet.Seq] = packet.Data[:packet.Size]

	tpl.Log("[recv data] %v (%v) %v", packet.Seq*tpl.PACKET_SIZE, len(packet.Data), getStatus(packet.Seq))
	var flag uint16 = 2
	if packet.Flags == 1 {
		finalPacketId = int(packet.Seq)
	}
	if packet.Flags == 1 && haveAllPackets(finalPacketId) {
		tpl.Log("recv final data packet")
		done = true
		flag = 3
		// TODO: add a final shutdown flag thing
	}

	var data [tpl.PACKET_SIZE]byte
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
