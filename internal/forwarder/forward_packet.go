package forwarder

import (
	"encoding/binary"
	"fmt"
	"net"
	"packetpath/internal/config"
	"packetpath/internal/header"
	"sync"

	"github.com/one-api/godivert"
)

type ConnKey struct {
	SrcIP   [4]byte
	SrcPort uint16
	DstIP   [4]byte
	DstPort uint16
}

var headerSent sync.Map

type PacketContext struct {
	IHL          int
	TcpHeaderLen int
	ReadLen      uint32
	Packet       []byte
	Addr         *godivert.Address
	Conn         *godivert.Divert
	Cfg          config.Config
}

func makeKey(packet []byte, ihl int) ConnKey {
	var k ConnKey
	copy(k.SrcIP[:], packet[12:16])
	k.SrcPort = binary.BigEndian.Uint16(packet[ihl : ihl+2])
	copy(k.DstIP[:], packet[16:20])
	k.DstPort = binary.BigEndian.Uint16(packet[ihl+2 : ihl+4])
	return k
}

func reverseKey(packet []byte, ihl int) ConnKey {
	var k ConnKey
	copy(k.SrcIP[:], packet[16:20])
	k.SrcPort = binary.BigEndian.Uint16(packet[ihl+2 : ihl+4])
	copy(k.DstIP[:], packet[12:16])
	k.DstPort = binary.BigEndian.Uint16(packet[ihl : ihl+2])
	return k
}

func processOutbound(
	ctx *PacketContext,
) {
	dstIp := ctx.Packet[16:20]
	ip := fmt.Sprintf("%d.%d.%d.%d", dstIp[0], dstIp[1], dstIp[2], dstIp[3])
	if ip != ctx.Cfg.RedirectIp {
		ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
		return
	}

	flags := ctx.Packet[ctx.IHL+13]
	isPSH := flags&0x08 != 0

	key := makeKey(ctx.Packet, ctx.IHL)
	_, alreadyHeaderized := headerSent.Load(key)

	if isPSH && !alreadyHeaderized {
		err := header.ApplyHeader(
			&ctx.Packet,
			net.ParseIP(ctx.Cfg.RedirectIp).To4(),
			ctx.Cfg.RedirectPort,
			ctx.IHL,
			ctx.TcpHeaderLen,
		)

		if err != nil {
			ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
			return
		}

		headerSent.Store(key, true)
	} else if alreadyHeaderized {
		seq := binary.BigEndian.Uint32(ctx.Packet[ctx.IHL+4 : ctx.IHL+8])
		seq += header.CustomHeaderSize
		binary.BigEndian.PutUint32(ctx.Packet[ctx.IHL+4:ctx.IHL+8], seq)
	}

	newIP := net.ParseIP(ctx.Cfg.RelayIp).To4()
	copy(ctx.Packet[16:20], newIP)
	binary.BigEndian.PutUint16(ctx.Packet[ctx.IHL+2:ctx.IHL+4], ctx.Cfg.RelayPort)

	newLen := len(ctx.Packet)
	godivert.CalcChecksums(ctx.Packet[:newLen], 0)
	ctx.Addr.SetTCPChecksum(false)
	ctx.Addr.SetIPChecksum(false)

	ctx.Conn.Send(ctx.Packet[:newLen], ctx.Addr)
}

func processInbound(
	ctx *PacketContext,
) {
	copy(ctx.Packet[12:16], net.ParseIP(ctx.Cfg.RedirectIp).To4())
	binary.BigEndian.PutUint16(ctx.Packet[ctx.IHL:ctx.IHL+2], ctx.Cfg.RedirectPort)

	key := reverseKey(ctx.Packet, ctx.IHL)
	_, wasHeaderized := headerSent.Load(key)
	if wasHeaderized {
		seq := binary.BigEndian.Uint32(ctx.Packet[ctx.IHL+8 : ctx.IHL+12])
		seq -= header.CustomHeaderSize
		binary.BigEndian.PutUint32(ctx.Packet[ctx.IHL+8:ctx.IHL+12], seq)
	}

	godivert.CalcChecksums(ctx.Packet[:ctx.ReadLen], 0)

	ctx.Addr.SetIPChecksum(false)
	ctx.Addr.SetTCPChecksum(false)

	ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
}

func Forward(cfg config.Config, readLen uint32, packet []byte, conn *godivert.Divert, addr *godivert.Address) {
	ihl := int(packet[0]&0x0f) * 4

	if ihl < 20 || int(readLen) < ihl {
		conn.Send(packet[:readLen], addr)
		return
	}

	tcpHeaderLen := int(packet[ihl+12]>>4) * 4

	if addr.Outbound() {
		ctx := &PacketContext{
			IHL:          ihl,
			TcpHeaderLen: tcpHeaderLen,
			ReadLen:      readLen,
			Packet:       packet,
			Addr:         addr,
			Conn:         conn,
			Cfg:          cfg,
		}
		processOutbound(ctx)
		return
	}

	srcIp := packet[12:16]
	ip := fmt.Sprintf("%d.%d.%d.%d", srcIp[0], srcIp[1], srcIp[2], srcIp[3])
	if ip == cfg.RelayIp {
		ctx := &PacketContext{
			IHL:          ihl,
			TcpHeaderLen: tcpHeaderLen,
			ReadLen:      readLen,
			Packet:       packet,
			Addr:         addr,
			Conn:         conn,
			Cfg:          cfg,
		}
		processInbound(ctx)
		return
	}

	conn.Send(packet[:readLen], addr)
}
