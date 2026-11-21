package header

import (
	"net"
	"testing"
)

func TestPrepareHeaderSlot(t *testing.T) {
	packet := make([]byte, 65535)
	packet = packet[:2]
	copy(packet[:2], []byte{1, 2})

	err := prepareHeaderSlot(&packet)
	if err != nil {
		t.Fatal(err)
	}

	get := len(packet)
	want := 8

	if get != want {
		t.Fatalf("want %d, got %d", want, get)
	}

	if cap(packet) != 65535 {
		t.Fatalf("want %d, got %d", 65535, cap(packet))
	}
}

func TestApplyHeader_WithAdditionalData(t *testing.T) {
	packet := make([]byte, 65535)

	dummyTraffic := []byte{
		69,
		0,
		0,
		52,
		24,
		42,
		64,
		0,
		128,
		6,
		0,
		0,
		192,
		168,
		110,
		79,
		111,
		111,
		11,
		111,
		42,
		233,
		9,
		19,
		207,
		158,
		59,
		252,
		0,
		0,
		0,
		0,
		128,
		2,
		255,
		255,
		187,
		162,
		0,
		0,
		2,
		4,
		5,
		180,
		1,
		3,
		3,
		8,
		1,
		1,
		4,
		2,
		99,
		99,
		99,
		99,
	} // last 4 byte with 99 value is fake data to simulate what will happen if packet has body.

	packet = append(packet[:0], dummyTraffic...)
	want := len(packet) + 6

	ihl := int(packet[0]&0x0f) * 4
	tcpHeaderLen := int(packet[32]>>4) * 4

	err := ApplyHeader(&packet, net.ParseIP("244.244.244.244").To4(), 65534, ihl, tcpHeaderLen)
	if err != nil {
		t.Fatal(err)
	}

	got := len(packet)
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestApplyHeader_WithoutAdditionalData(t *testing.T) {
	packet := make([]byte, 65535)

	dummyTraffic := []byte{
		69,
		0,
		0,
		52,
		24,
		42,
		64,
		0,
		128,
		6,
		0,
		0,
		192,
		168,
		110,
		79,
		111,
		111,
		11,
		111,
		42,
		233,
		9,
		19,
		207,
		158,
		59,
		252,
		0,
		0,
		0,
		0,
		128,
		2,
		255,
		255,
		187,
		162,
		0,
		0,
		2,
		4,
		5,
		180,
		1,
		3,
		3,
		8,
		1,
		1,
		4,
		2,
	}

	packet = append(packet[:0], dummyTraffic...)
	want := len(packet) + 6

	ihl := int(packet[0]&0x0f) * 4
	tcpHeaderLen := int(packet[32]>>4) * 4

	err := ApplyHeader(&packet, net.ParseIP("244.244.244.244").To4(), 65534, ihl, tcpHeaderLen)
	if err != nil {
		t.Fatal(err)
	}

	got := len(packet)
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestCalcHeaderNumber_SeqNum(t *testing.T) {
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0} // tcp.seq and tcp.seq_raw are 4-bytes (uint32)
	CalcHeaderNumber(p, 0, CalculateSeqNum)

	want := 6
	got := int(p[len(p)-1])

	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestCalcHeaderNumber_AckNum(t *testing.T) {
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12} // tcp.ack and tcp.ack_raw are 4-bytes (uint32)
	CalcHeaderNumber(p, 0, CalculateAckNum)
	t.Log(p)

	want := 6
	got := int(p[len(p)-1])

	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}
