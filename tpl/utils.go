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
	buf := make([]byte, 2048)
	size, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println(err)
	}
	if size > 1 {
		fmt.Println("not empty packet")
	}

	bufReader := bytes.NewReader(buf)
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
