package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"syscall"

	"github.com/one-api/godivert"
	"github.com/one-api/godivert/types"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/windows"
)

var (
	flowMap sync.Map
)

func makeFlowKey(protocol uint8, srcIp string, srcPort uint16, dstIp string, dstPort uint16) string {
	return fmt.Sprintf("%v|%v|%v|%v|%v", protocol, srcIp, srcPort, dstIp, dstPort)
}

func listenFlows(divert *godivert.Divert) {
	fmt.Println("[+] Running flow listener...")

	packet := make([]byte, 2048)
	var addr godivert.Address

	for {
		_, err := divert.Recv(packet, &addr)
		if err != nil {
			log.Printf("Error receiving packet: %s", err)
			continue
		}

		flowData := addr.Flow()
		if flowData == nil || flowData.Protocol != 6 {
			continue
		}

		if addr.IPv6() {
			continue
		}

		processName := getProcessName(flowData.ProcessId)

		if processName != "AION2.exe" {
			continue
		}

		localAddr := Ipv4ToString(flowData.LocalAddr[0])
		localPort := flowData.LocalPort
		remoteAddr := Ipv4ToString(flowData.RemoteAddr[0])
		remotePort := flowData.RemotePort

		if localAddr == remoteAddr {
			continue
		}

		if remotePort == 443 {
			continue
		}

		key := makeFlowKey(6, localAddr, localPort, remoteAddr, remotePort)

		if addr.Event == uint8(types.EventFlowEstablished) {
			flowMap.Store(key, processName)
		} else if addr.Event == uint8(types.EventFlowDeleted) {
			flowMap.Delete(key)
		}
	}
}

func main() {
	flowHandle, err := godivert.New("tcp", godivert.LayerFlow, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer flowHandle.Close()
	go listenFlows(flowHandle)

	netHandle, err := godivert.New("outbound and tcp", godivert.LayerNetwork, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer netHandle.Close()

	packet := make([]byte, 65535)
	var addr godivert.Address

	for {
		readLen, err := netHandle.Recv(packet, &addr)
		if err != nil {
			log.Printf("Error receiving packet: %s", err)
			continue
		}

		showPacketInfo(packet[:readLen])

		netHandle.Send(packet[:readLen], &addr)
	}
}

func showPacketInfo(packet []byte) {
	iph, err := ipv4.ParseHeader(packet)
	if err != nil {
		fmt.Printf("Error parsing IPv4: %v\n", err)
		return
	}

	payload := packet[iph.Len:]

	if iph.Protocol != 6 {
		return
	}

	var srcPort uint16
	var dstPort uint16

	if len(payload) >= 20 {
		srcPort = binary.BigEndian.Uint16(payload[0:2])
		dstPort = binary.BigEndian.Uint16(payload[2:4])
	}

	key := makeFlowKey(6, iph.Src.String(), srcPort, iph.Dst.String(), dstPort)

	if _, ok := flowMap.Load(key); ok {
		fmt.Println(key)
	}
}

func getProcessName(pid uint32) string {
	if pid == 0 {
		return "System (Idle)"
	}
	if pid == 4 {
		return "System"
	}

	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "unknown"
	}
	defer windows.CloseHandle(h)

	var size uint32 = 1024
	buf := make([]uint16, size)
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "unknown"
	}

	path := syscall.UTF16ToString(buf[:size])
	// Extract base name
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func Ipv4ToString(ip uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}
