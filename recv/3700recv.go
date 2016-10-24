package main

import (
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
	var startWaitingFinalAck = time.Now()
	for !ackDone {
		if time.Since(startWaitingFinalAck).Seconds() > 4 {
			ackDone = true
			break
		}

		conn.SetDeadline(time.Now().Add(time.Second * 1))
		packet, retAddr := tpl.ReadPacket(conn)
		if packet.Flags != 4 {

			var data []byte
			teardown := tpl.Packet{
				Seq:   packet.Seq,
				Flags: 3,
				Data:  data,
			}

			buf := tpl.WriteBytes(teardown)

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
	dataChunks[packet.Seq] = packet.Data

	tpl.Log("[recv data] %v (%v) %v", packet.Seq*tpl.PACKET_SIZE, len(packet.Data), getStatus(packet.Seq))
	var flag uint32 = 2
	if packet.Flags == 1 {
		finalPacketId = int(packet.Seq)
	}
	if haveAllPackets(finalPacketId) {
		tpl.Log("recv final data packet")
		done = true
		flag = 3
		// TODO: add a final shutdown flag thing
	}

	var data []byte
	// send an acknowledgement packet
	acket := tpl.Packet{
		Seq:   packet.Seq,
		Flags: flag,
		Data:  data,
	}

	buf := tpl.WriteBytes(acket)

	conn.WriteTo(buf.Bytes(), retAddr)

	// the issue with the last packet is it could be dropped during delivery

	// so we might actually need a 3 way handshake or something around closing out
}
