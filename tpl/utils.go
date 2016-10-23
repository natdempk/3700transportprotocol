package tpl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func ReadPacket(conn net.UDPConn) (packet Packet) {
	var buf []byte
	_, _, _ = conn.ReadFromUDP(buf)

	bufReader := bytes.NewReader(buf)

	// Read into an empty struct.
	packet = Packet{}
	_ = binary.Read(bufReader, binary.LittleEndian, &packet)
	return
}

func Log(format string, a ...interface{}) {
	timestampFormat := "%d " + format + "\n"
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
