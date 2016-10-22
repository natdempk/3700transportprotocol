package tpl

type Packet struct {
	Seq       uint32 // sequence number
	Ack       uint32 // last contiguous packet ID seen
	AdvWindow uint16 // advertised window
	Flags     uint16 // 1 = done, 2 = ack, 3 = final
	Data      []byte
}

const PACKET_SIZE = 10
