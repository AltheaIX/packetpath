package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
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

			header := make([]byte, 6)
			if _, err := io.ReadFull(c, header); err != nil {
				log.Printf("Failed to read header: %v\n", err)
				return
			}

			origIp := net.IP(header[0:4])
			origPort := binary.BigEndian.Uint16(header[4:6])
			fmt.Printf("Relaying %s <-> %s\n", c.RemoteAddr(), origIp)

			server, err := net.Dial("tcp", fmt.Sprintf("%s:%d", origIp.String(), origPort))
			if err != nil {
				log.Printf("Failed to dial upstream %s:%d: %v\n", origIp.String(), origPort, err)
				return
			}
			defer server.Close()

			go io.Copy(server, c)
			io.Copy(c, server)
		}(conn)
	}
}
