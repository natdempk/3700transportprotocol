package main

type Packet struct {
	Seq       uint32
	Ack       uint32
	AdvWindow uint16
	Flags     uint16
	Data      []byte
}
