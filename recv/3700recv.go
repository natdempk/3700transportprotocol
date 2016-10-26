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
var ignoreLast = false

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

	var output []byte
	for i := 0; i < len(dataChunks); i++ {
		output = append(output, dataChunks[uint32(i)]...)
	}
	output = tpl.DecompressBytes(output, ignoreLast)
	// print out final received data
	fmt.Printf("%s", output)

	// Wait up to 4 seconds to recieve a final ackDone
	var ackDone = false
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
	tpl.Log("[recv data] %v (%v) %v", packet.Seq*tpl.PACKET_SIZE, len(packet.Data), getStatus(packet.Seq))

	dataChunks[packet.Seq] = packet.Data
	var flag uint8 = 2
	if packet.Flags == 1 || packet.Flags == 5 {
		finalPacketId = int(packet.Seq)
		if packet.Flags == 5 {
			ignoreLast = true
		}
	}

	if haveAllPackets(finalPacketId) {
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
	tpl.Log("[send ack], %v %v", acket.Seq, acket.Flags)
	conn.WriteTo(buf.Bytes(), retAddr)
}
