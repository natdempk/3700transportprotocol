package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"../tpl"
)

var dataChunks = make(map[uint32][]byte)
var done = false
var finalPacketId = -1
var conn net.PacketConn

func main() {
	conn, _ = net.ListenUDP("udp", nil)
	// [::]:portNumer, the above returns a random open port
	port, _ := strconv.Atoi(conn.LocalAddr().String()[5:])
	defer conn.Close()
	tpl.Log("[bound] %d\n", port)

	for !done {
		packet, retAddr := tpl.ReadPacket(conn)
		handleConnection(packet, retAddr)
	}

	for i := 0; i < len(dataChunks); i++ {
		fmt.Printf("%s", dataChunks[uint32(i)])
	}

	// Wait up to 4 seconds to recieve a final ackDone
	var ackDone bool = false
	var startWaitingFinalAck = time.Now()
	for !ackDone {
		if time.Since(startWaitingFinalAck).Seconds() > 2 {
			ackDone = true
			break
		}

		conn.SetDeadline(time.Now().Add(time.Second * 1))
		packet, retAddr := tpl.ReadPacket(conn)
		// If it's not final ack, resend we're done
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
	if _, ok := dataChunks[seq]; ok {
		return "IGNORED"
	}
	for i := uint32(0); i < seq; i++ {
		if _, ok := dataChunks[seq]; !ok {
			return "ACCEPTED (out-of-order)"
		}
	}
	return "ACCEPTED (in-order)"
}

func handleConnection(packet tpl.Packet, retAddr net.Addr) {
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
	}

	var data []byte
	acket := tpl.Packet{
		Seq:   packet.Seq,
		Flags: flag,
		Data:  data,
	}

	buf := tpl.WriteBytes(acket)
	conn.WriteTo(buf.Bytes(), retAddr)
}
