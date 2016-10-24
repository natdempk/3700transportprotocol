package tpl

type Packet struct {
	Seq       uint32 // sequence number
	Size      uint32
	Ack       uint32 // last contiguous packet ID seen
	AdvWindow uint32 // advertised window
	Flags     uint32 // 1 = done, 2 = ack, 3 = final, 4 = ack final
	Data      []byte
}

const PACKET_SIZE = 1024 * 4
