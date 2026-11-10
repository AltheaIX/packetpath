package main

import (
	"fmt"
	"log"

	"github.com/one-api/godivert"
)

func main() {
	divert, err := godivert.New("true", godivert.LayerFlow, 0, 0)
	if err != nil {
		log.Fatal("Failed to open WinDivert (run as Administrator):", err)
	}
	defer divert.Close()

	fmt.Println("Intercepting TCP packets... Press Ctrl+C to stop.")

	packet := make([]byte, 2048)
	var addr godivert.Address

	for {
		_, err = divert.Recv(packet, &addr)
		if err != nil {
			log.Printf("Error receiving packet: %s", err)
			continue
		}

		flowData := addr.Flow()
		if flowData == nil {
			continue
		}

		if flowData.Protocol != 6 {
			continue
		}

		fmt.Println(flowData)
	}
}

func Ipv4ToString(ip uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}
