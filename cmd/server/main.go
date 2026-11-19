package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	targetSrvPtr := flag.String("target_server", "127.0.0.1:8080", "original destination of traffic")
	targetSrv := *targetSrvPtr

	listener, err := net.Listen("tcp", ":2323")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("Listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			server, err := net.Dial("tcp", targetSrv)
			if err != nil {
				log.Printf("Failed to connect to server: %v\n", err)
				return
			}
			defer server.Close()

			log.Printf("Relaying %s <-> %s\n", c.RemoteAddr(), targetSrv)

			go io.Copy(server, c)
			io.Copy(c, server)
		}(conn)
	}
}
