package main

import (
	"flag"
	"fmt"
	"log"
	"packetpath/internal/config"
	"packetpath/internal/forwarder"
	"packetpath/internal/interceptor"
	"sync"

	"github.com/one-api/godivert"
	"github.com/one-api/godivert/types"
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

		processName := interceptor.GetProcessName(flowData.ProcessId)

		if processName != "curl.exe" {
			continue
		}

		localAddr := interceptor.Ipv4ToString(flowData.LocalAddr[0])
		localPort := flowData.LocalPort
		remoteAddr := interceptor.Ipv4ToString(flowData.RemoteAddr[0])
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
	redirectIP := flag.String("redirect_ip", "127.0.0.1", "ip of traffic to be redirected")
	redirectPortRaw := flag.String("redirect_port", "8080", "port of traffic to be redirected")
	relayIP := flag.String("relay_ip", "127.0.0.1", "relay ip")
	relayPortRaw := flag.String("relay_port", "8080", "relay port")
	flag.Parse()

	redirectPort, err := config.ParsePort(*redirectPortRaw)
	if err != nil {
		log.Fatal(err)
	}

	relayPort, err := config.ParsePort(*relayPortRaw)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Config{
		RedirectIp:   *redirectIP,
		RedirectPort: redirectPort,
		RelayIp:      *relayIP,
		RelayPort:    relayPort,
	}

	fmt.Printf("Traffic From %s will be Redirected to %s\n", cfg.RedirectIp, cfg.RelayIp)

	flowHandle, err := godivert.New("tcp", godivert.LayerFlow, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer flowHandle.Close()
	go listenFlows(flowHandle)

	filter := "tcp"
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

		forwarder.Forward(cfg, readLen, packet[:readLen], netHandle, &addr)
	}
}
