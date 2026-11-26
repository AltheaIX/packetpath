package main

import (
	"flag"
	"fmt"
	"log"
	"packetpath/internal/config"
	"packetpath/internal/forwarder"
	"packetpath/internal/header"
	"packetpath/internal/interceptor"
	"sync"
	"time"
)

func mapCounterHelper(Map *sync.Map) int {
	mapLength := 0

	Map.Range(
		func(key, value any) bool {
			mapLength++
			return true
		},
	)

	return mapLength
}

func startMapWatcher() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("[+] Map Watcher is on")

	for {
		select {
		case <-ticker.C:
			socketMapLength := mapCounterHelper(&interceptor.SocketProgramMap)
			headerSentLength := mapCounterHelper(&header.HeaderSent)
			originTargetMapLength := mapCounterHelper(&forwarder.OriginTargetMap)

			fmt.Printf(
				"socketMapLength=%d, headerSentLength=%d, originTargetMapLength=%d\n",
				socketMapLength,
				headerSentLength,
				originTargetMapLength,
			)
		}
	}
}

func main() {
	relayIP := flag.String("relay_ip", "127.0.0.1", "relay ip")
	relayPortRaw := flag.String("relay_port", "8080", "relay port")
	programName := flag.String("program", "", "program name with .exe")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	if *debug {
		go startMapWatcher()
	}

	relayPort, err := config.ParsePort(*relayPortRaw)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Config{
		ProgramName: *programName,
		RelayIp:     *relayIP,
		RelayPort:   relayPort,
	}

	fmt.Printf("Traffic From %s will be Redirected to %s\n", cfg.ProgramName, cfg.RelayIp)

	go interceptor.StartInterceptor(cfg)
	forwarder.StartForwarder(cfg)
}
