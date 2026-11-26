package forwarder

import (
	"encoding/binary"
	"fmt"
	"net"
	"packetpath/internal/config"
	"packetpath/internal/header"
	"packetpath/internal/interceptor"

	"github.com/one-api/godivert"
)

type PacketContext struct {
	IHL          int
	TcpHeaderLen int
	ReadLen      uint32
	Packet       []byte
	Addr         *godivert.Address
	Conn         *godivert.Divert
	Cfg          config.Config
}

func processOutbound(
	ctx *PacketContext,
) {
	srcIP := ctx.Packet[12:16]
	srcPort := binary.BigEndian.Uint16(ctx.Packet[ctx.IHL : ctx.IHL+2])

	dstIP := ctx.Packet[16:20]
	dstPort := binary.BigEndian.Uint16(ctx.Packet[ctx.IHL+2 : ctx.IHL+4])

	srcIPStr := fmt.Sprintf("%d.%d.%d.%d", srcIP[0], srcIP[1], srcIP[2], srcIP[3])
	dstIPStr := fmt.Sprintf("%d.%d.%d.%d", dstIP[0], dstIP[1], dstIP[2], dstIP[3])

	socketKey := interceptor.MakeSocketProgramKey(srcIPStr, srcPort, dstIPStr, dstPort)
	if interceptor.ResolveProgramName(socketKey) == "unknown" {
		ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
		return
	}

	dstIpCopy := make([]byte, len(dstIP))
	copy(dstIpCopy, dstIP)
	orig := MakeOriginTargetKey(dstIpCopy, dstPort)

	OriginTargetMap.Store(srcPort, orig)

	flags := ctx.Packet[ctx.IHL+13]
	isPSH := flags&0x08 != 0

	key := header.MakeKey(ctx.Packet, ctx.IHL, header.OutboundFlag)
	_, alreadyHeaderized := header.HeaderSent.Load(key)

	if isPSH && !alreadyHeaderized {
		err := header.ApplyHeader(
			&ctx.Packet,
			dstIpCopy,
			dstPort,
			ctx.IHL,
			ctx.TcpHeaderLen,
		)

		if err != nil {
			ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
			return
		}

		header.HeaderSent.Store(key, true)
	} else if alreadyHeaderized {
		header.CalcHeaderNumber(ctx.Packet, ctx.IHL, header.CalculateSeqNum)
	}

	isRST := flags&0x04 != 0
	isFIN := flags&0x01 != 0

	if isRST || isFIN {
		header.HeaderSent.Delete(key)
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
	srcIp := ctx.Packet[12:16]
	ip := fmt.Sprintf("%d.%d.%d.%d", srcIp[0], srcIp[1], srcIp[2], srcIp[3])
	if ip != ctx.Cfg.RelayIp {
		ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
		return
	}

	dstPort := binary.BigEndian.Uint16(ctx.Packet[ctx.IHL+2 : ctx.IHL+4])
	val, ok := OriginTargetMap.Load(dstPort)
	if !ok {
		ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
		return
	}
	orig := val.(OriginTargetKey)

	copy(ctx.Packet[12:16], orig.Ip)
	binary.BigEndian.PutUint16(ctx.Packet[ctx.IHL:ctx.IHL+2], orig.Port)

	key := header.MakeKey(ctx.Packet, ctx.IHL, header.InboundFlag)
	_, wasHeaderized := header.HeaderSent.Load(key)
	if wasHeaderized {
		header.CalcHeaderNumber(ctx.Packet, ctx.IHL, header.CalculateAckNum)
	}

	godivert.CalcChecksums(ctx.Packet[:ctx.ReadLen], 0)

	ctx.Addr.SetIPChecksum(false)
	ctx.Addr.SetTCPChecksum(false)

	ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
}

func forward(cfg config.Config, readLen uint32, packet []byte, conn *godivert.Divert, addr *godivert.Address) {
	ihl := int(packet[0]&0x0f) * 4

	if ihl < 20 || int(readLen) < ihl {
		conn.Send(packet[:readLen], addr)
		return
	}

	tcpHeaderLen := int(packet[ihl+12]>>4) * 4

	ctx := &PacketContext{
		IHL:          ihl,
		TcpHeaderLen: tcpHeaderLen,
		ReadLen:      readLen,
		Packet:       packet,
		Addr:         addr,
		Conn:         conn,
		Cfg:          cfg,
	}

	if addr.Outbound() {
		processOutbound(ctx)
		return
	}

	processInbound(ctx)
}
