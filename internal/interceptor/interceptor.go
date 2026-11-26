package interceptor

import (
	"log"
	"packetpath/internal/config"

	"github.com/one-api/godivert"
)

func StartInterceptor(cfg config.Config) {
	socketHandle, err := godivert.New(
		"tcp",
		godivert.LayerSocket,
		0,
		godivert.FlagSniff|godivert.FlagRecvOnly,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer socketHandle.Close()

	packet := make([]byte, 1)
	var addr godivert.Address

	for {
		_, err := socketHandle.Recv(packet, &addr)
		if err != nil {
			log.Printf("Socket Recv error: %v", err)
			return
		}

		sockData := addr.Socket()
		if sockData == nil {
			return
		}

		intercept(sockData, cfg.ProgramName)
	}
}
