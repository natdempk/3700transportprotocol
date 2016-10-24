package tpl

type Packet struct {
	Seq   uint32 // sequence number
	Flags uint8  // 1 = done, 2 = ack, 3 = final, 4 = ack final
	Data  []byte
}

const PACKET_SIZE = 1024 * 32
