package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/one-api/godivert"
	"github.com/one-api/godivert/types"
	"golang.org/x/sys/windows"
)

var (
	flowMap    sync.Map
	reverseMap sync.Map
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

		if processName != "curl.exe" {
			continue
		}

		localAddr := Ipv4ToString(flowData.LocalAddr[0])
		localPort := flowData.LocalPort
		remoteAddr := Ipv4ToString(flowData.RemoteAddr[0])
		remotePort := flowData.RemotePort

		if localAddr == remoteAddr {
			continue
		}

		//if remotePort == 443 {
		//	continue
		//}

		key := makeFlowKey(6, localAddr, localPort, remoteAddr, remotePort)

		if addr.Event == uint8(types.EventFlowEstablished) {
			flowMap.Store(key, processName)
		} else if addr.Event == uint8(types.EventFlowDeleted) {
			flowMap.Delete(key)
		}
	}
}

func main() {
	redirectIpPtr := flag.String("redirect_ip", "127.0.0.1", "ip of traffic to be redirected")
	relayIpPtr := flag.String("ip", "127.0.0.1", "relay ip")
	flag.Parse()

	relayIp := *relayIpPtr
	redirectIp := *redirectIpPtr
	fmt.Printf("Redirecting Traffic %s from Traffic Redirected to %s\n", redirectIp, relayIp)

	flowHandle, err := godivert.New("tcp", godivert.LayerFlow, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer flowHandle.Close()
	go listenFlows(flowHandle)

	filter := "!loopback and tcp"
	netHandle, err := godivert.New(filter, godivert.LayerNetwork, 0, 0)
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

		if readLen < 20 {
			netHandle.Send(packet[:readLen], &addr)
			continue
		}

		if addr.Outbound() {
			dstIp := packet[16:20]
			ip := fmt.Sprintf("%d.%d.%d.%d", dstIp[0], dstIp[1], dstIp[2], dstIp[3])
			if ip == redirectIp {
				fmt.Println("[+] Outbound", packet[:readLen])
				newIP := net.ParseIP(relayIp).To4()
				copy(packet[16:20], newIP)
				godivert.CalcChecksums(packet[:readLen], 0)
				fmt.Println("[+] Modified Outbound", packet[:readLen])

				addr.SetTCPChecksum(false)
				addr.SetIPChecksum(false)
			}

			netHandle.Send(packet[:readLen], &addr)
			continue
		}

		//srcIp := packet[12:16]
		//ip := fmt.Sprintf("%d.%d.%d.%d", srcIp[0], srcIp[1], srcIp[2], srcIp[3])
		//if ip == "" {
		//	fmt.Println("[+] Inbound", packet[:readLen])
		//
		//	originalServerIp := []byte{}
		//	copy(packet[12:16], originalServerIp)
		//
		//	godivert.CalcChecksums(packet[:readLen], 0)
		//
		//	fmt.Println("[+] Modified Inbound (Spoofed as Original Server):", packet[:readLen])
		//}

		netHandle.Send(packet[:readLen], &addr)
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
