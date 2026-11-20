package forwarder

import (
	"encoding/binary"
	"fmt"
	"net"
	"packetpath/internal/config"

	"github.com/one-api/godivert"
)

type PacketContext struct {
	IHL     int
	ReadLen uint32
	Packet  []byte
	Addr    *godivert.Address
	Conn    *godivert.Divert
	Cfg     config.Config
}

func processOutbound(
	ctx *PacketContext,
) {
	dstIp := ctx.Packet[16:20]
	ip := fmt.Sprintf("%d.%d.%d.%d", dstIp[0], dstIp[1], dstIp[2], dstIp[3])
	if ip == ctx.Cfg.RedirectIp {
		newIP := net.ParseIP(ctx.Cfg.RelayIp).To4()
		copy(ctx.Packet[16:20], newIP)
		binary.BigEndian.PutUint16(ctx.Packet[ctx.IHL+2:ctx.IHL+4], ctx.Cfg.RelayPort)
		godivert.CalcChecksums(ctx.Packet[:ctx.ReadLen], 0)

		ctx.Addr.SetTCPChecksum(false)
		ctx.Addr.SetIPChecksum(false)
	}

	ctx.Conn.Send(ctx.Packet[:ctx.ReadLen], ctx.Addr)
}

func processInbound(
	ctx *PacketContext,
) {
	copy(ctx.Packet[12:16], net.ParseIP(ctx.Cfg.RedirectIp).To4())
	binary.BigEndian.PutUint16(ctx.Packet[ctx.IHL:ctx.IHL+2], ctx.Cfg.RedirectPort)
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

	if addr.Outbound() {
		ctx := &PacketContext{IHL: ihl, ReadLen: readLen, Packet: packet, Addr: addr, Conn: conn, Cfg: cfg}
		processOutbound(ctx)
		return
	}

	srcIp := packet[12:16]
	ip := fmt.Sprintf("%d.%d.%d.%d", srcIp[0], srcIp[1], srcIp[2], srcIp[3])
	if ip == cfg.RelayIp {
		ctx := &PacketContext{IHL: ihl, ReadLen: readLen, Packet: packet, Addr: addr, Conn: conn, Cfg: cfg}
		processInbound(ctx)
		return
	}

	conn.Send(packet[:readLen], addr)
}
