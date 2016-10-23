package tpl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func ReadPacketC(conn net.Conn) (packet Packet) {
	buf := make([]byte, PACKET_SIZE+100)
	conn.Read(buf)

	bufReader := bytes.NewReader(buf)
	packet = Packet{}
	_ = binary.Read(bufReader, binary.LittleEndian, &packet)
	return
}

func ReadPacket(conn net.PacketConn) (packet Packet, fromAddr net.Addr) {
	buf := make([]byte, PACKET_SIZE+100)
	_, fromAddr, err := conn.ReadFrom(buf)
	if err != nil {
		fmt.Println(err)
	}

	bufReader := bytes.NewReader(buf)
	packet = Packet{}
	_ = binary.Read(bufReader, binary.LittleEndian, &packet)
	return
}

func Log(format string, a ...interface{}) {
	timestampFormat := "%s " + format + "\n"
	a = append([]interface{}{Timestamp()}, a...)
	fmt.Fprintf(os.Stderr, timestampFormat, a...)
}

func Timestamp() (timestamp string) {
	nanos := time.Now().UnixNano()
	micros := nanos / 1000
	timestamp = fmt.Sprintf("%d", micros)
	return
}

func Min(a, b int) int {
	if a < b {
		return a

	}
	return b
}
