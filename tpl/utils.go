package tpl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func WriteBytes(packet Packet) (buff bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Seq)
	binary.Write(buf, binary.LittleEndian, packet.Flags)

	for i := 0; i < len(packet.Data); i++ {
		binary.Write(buf, binary.LittleEndian, packet.Data[i])
	}
	buff = *buf
	return
}

func ReadBytes(buff *bytes.Reader) (packet Packet) {
	packet = Packet{}
	binary.Read(buff, binary.LittleEndian, &packet.Seq)
	binary.Read(buff, binary.LittleEndian, &packet.Flags)

	var data []byte
	for {
		newByte, err := buff.ReadByte()
		if err == io.EOF {
			packet.Data = data
			return
		}
		data = append(data, newByte)
	}
	return

}

func ReadPacketC(conn net.Conn) (packet Packet) {
	buf := make([]byte, PACKET_SIZE+100)
	size, _ := conn.Read(buf)

	bufReader := bytes.NewReader(buf[:size])
	packet = ReadBytes(bufReader)
	return
}

func ReadPacket(conn net.PacketConn) (packet Packet, fromAddr net.Addr) {
	buf := make([]byte, PACKET_SIZE+100)
	size, fromAddrp, _ := conn.ReadFrom(buf)

	bufReader := bytes.NewReader(buf[:size])
	packet = ReadBytes(bufReader)
	fromAddr = fromAddrp
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
