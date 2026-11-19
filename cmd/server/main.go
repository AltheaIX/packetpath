package main

import (
	"fmt"
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
		log.Printf("connection from %s", conn.RemoteAddr())

		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 4096)
			for {
				n, err := c.Read(buf)
				if n > 0 {
					log.Printf("read %d bytes: %x", n, buf[:n]) // hex, not string
				}
				if err != nil {
					log.Printf("read error: %v", err)
					return
				}
			}
		}(conn)
	}
}
