package header

import (
	"encoding/binary"
	"sync"
)

const (
	OutboundFlag = uint8(0x01)
	InboundFlag  = uint8(0x02)
)

// HeaderSent is a map to make sure 6-byte custom header only sent once on Outbound.
// Also, for Inbound logic whether to re-calculate ACK number or not.
var HeaderSent sync.Map

type SentKey struct {
	SrcIP   [4]byte
	SrcPort uint16
	DstIP   [4]byte
	DstPort uint16
}

func MakeKey(packet []byte, ihl int, flag uint8) SentKey {
	var k SentKey

	if (flag & OutboundFlag) != 0 {
		copy(k.SrcIP[:], packet[12:16])
		k.SrcPort = binary.BigEndian.Uint16(packet[ihl : ihl+2])
		copy(k.DstIP[:], packet[16:20])
		k.DstPort = binary.BigEndian.Uint16(packet[ihl+2 : ihl+4])
	} else if (flag & InboundFlag) != 0 {
		copy(k.SrcIP[:], packet[16:20])
		k.SrcPort = binary.BigEndian.Uint16(packet[ihl+2 : ihl+4])
		copy(k.DstIP[:], packet[12:16])
		k.DstPort = binary.BigEndian.Uint16(packet[ihl : ihl+2])
	}

	return k
}
