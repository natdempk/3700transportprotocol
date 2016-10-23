package tpl

type Packet struct {
	Seq       uint32 // sequence number
	Size      uint16
	Ack       uint32 // last contiguous packet ID seen
	AdvWindow uint16 // advertised window
	Flags     uint16 // 1 = done, 2 = ack, 3 = final
	Data      [1024]byte
}

const PACKET_SIZE = 1024
