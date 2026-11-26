package forwarder

import (
	"log"
	"packetpath/internal/config"

	"github.com/one-api/godivert"
)

func StartForwarder(cfg config.Config) {
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

		forward(cfg, readLen, packet[:readLen], netHandle, &addr)
	}
}
