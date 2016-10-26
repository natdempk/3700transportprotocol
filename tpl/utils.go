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

func CompressBytes(data []byte) (compressed []byte, halfFull bool) {
	halfFull = false
	var word byte
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case byte('0'):
			word |= 0
		case byte('1'):
			word |= 1
		case byte('2'):
			word |= 2
		case byte('3'):
			word |= 3
		case byte('4'):
			word |= 4
		case byte('5'):
			word |= 5
		case byte('6'):
			word |= 6
		case byte('7'):
			word |= 7
		case byte('8'):
			word |= 8
		case byte('9'):
			word |= 9
		case byte('a'):
			word |= 10
		case byte('b'):
			word |= 11
		case byte('c'):
			word |= 12
		case byte('d'):
			word |= 13
		case byte('e'):
			word |= 14
		case byte('f'):
			word |= 15
		case byte('\n'):
			continue
		default:
			panic("okay")
		}
		if halfFull {
			compressed = append(compressed, word)
		}
		halfFull = !halfFull
		word = word << 4
	}
	if halfFull {
		compressed = append(compressed, word)
	}
	return
}

func DecompressBytes(data []byte, ignoreLastHalf bool) (decompressed []byte) {
	for i := 0; i < len(data); i++ {
		decompressed = append(decompressed, decompressHalfWord(data[i]>>4))

		decompressed = append(decompressed, decompressHalfWord(data[i]&15))
		if (i+1)%30 == 0 && i != 0 {
			decompressed = append(decompressed, byte('\n'))
		}
	}
	if ignoreLastHalf {
		decompressed = decompressed[:len(decompressed)-1]
	}
	return
}

func decompressHalfWord(halfWord byte) byte {
	switch halfWord {
	case 0:
		return byte('0')
	case 1:
		return byte('1')
	case 2:
		return byte('2')
	case 3:
		return byte('3')
	case 4:
		return byte('4')
	case 5:
		return byte('5')
	case 6:
		return byte('6')
	case 7:
		return byte('7')
	case 8:
		return byte('8')
	case 9:
		return byte('9')
	case 10:
		return byte('a')
	case 11:
		return byte('b')
	case 12:
		return byte('c')
	case 13:
		return byte('d')
	case 14:
		return byte('e')
	case 15:
		return byte('f')
	}
	return byte('z')
}

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
	err := binary.Read(buff, binary.LittleEndian, &packet.Seq)
	err = binary.Read(buff, binary.LittleEndian, &packet.Flags)

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

func ReadPacketC(conn net.Conn) (packet Packet, err error) {
	buf := make([]byte, PACKET_SIZE+100)
	size, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		// other side has already torn down the connection, so just pass the error on
		return packet, err
	}

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
