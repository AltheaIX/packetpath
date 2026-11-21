package header

import (
	"encoding/binary"
	"fmt"
)

const (
	CalculateSeqNum = uint8(0x01)
	CalculateAckNum = uint8(0x02)
)

const (
	CustomHeaderSize = 6
)

func prepareHeaderSlot(packet *[]byte) error {
	s := *packet
	newLen := CustomHeaderSize + len(s)
	if newLen < 0 || newLen > cap(s) {
		return fmt.Errorf("packet buffer overflow: need %d, cap %d", newLen, cap(s))
	}

	*packet = s[:newLen]
	return nil
}

func shiftCurrentFrame(packet []byte, ihl int, tcpHeaderLen int) error {
	if len(packet) < (ihl + tcpHeaderLen + CustomHeaderSize) {
		return fmt.Errorf("packet buffer overflow: apply prepareHeaderSlot() first")
	}

	copy(packet[ihl+tcpHeaderLen+CustomHeaderSize:], packet[ihl+tcpHeaderLen:])
	return nil
}

func fillCustomHeader(packet []byte, ip []byte, port uint16, ihl int, tcpHeaderLen int) error {
	copy(packet[ihl+tcpHeaderLen:ihl+tcpHeaderLen+4], ip)
	binary.BigEndian.PutUint16(packet[ihl+tcpHeaderLen+4:ihl+tcpHeaderLen+6], port)
	return nil
}

func modifyIpHeaderTotalLength(packet []byte) error {
	var newLen = uint16(len(packet))
	binary.BigEndian.PutUint16(packet[2:4], newLen)

	return nil
}

func ApplyHeader(packet *[]byte, ip []byte, port uint16, ihl int, tcpHeaderLen int) error {
	err := prepareHeaderSlot(packet)
	if err != nil {
		return err
	}

	err = shiftCurrentFrame(*packet, ihl, tcpHeaderLen)
	if err != nil {
		return err
	}

	err = fillCustomHeader(*packet, ip, port, ihl, tcpHeaderLen)
	if err != nil {
		return err
	}

	err = modifyIpHeaderTotalLength(*packet)
	if err != nil {
		return err
	}

	return nil
}

// CalcHeaderNumber Helper to calculate and rewrite SeqNum or AckNum into packet after rewriting data frame for 6-byte custom header.
func CalcHeaderNumber(packet []byte, ihl int, flag uint8) {
	if (flag & CalculateSeqNum) != 0 {
		seq := binary.BigEndian.Uint32(packet[ihl+4 : ihl+8])
		seq += CustomHeaderSize
		binary.BigEndian.PutUint32(packet[ihl+4:ihl+8], seq)
	} else if (flag & CalculateAckNum) != 0 {
		ack := binary.BigEndian.Uint32(packet[ihl+8 : ihl+12])
		ack -= CustomHeaderSize
		binary.BigEndian.PutUint32(packet[ihl+8:ihl+12], ack)
	}
}
