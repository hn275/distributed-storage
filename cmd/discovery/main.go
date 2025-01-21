package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hn275/dist-db/internal"
)

func main() {
	soc, err := net.Listen("tcp", ":"+internal.MustEnv("SOCKET_PORT"))
	if err != nil {
		panic(err)
	}
	defer soc.Close()

	log.Println("Local socket address", soc.Addr())
	for {
		conn, err := soc.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Println("Connection established:", conn.LocalAddr())
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	log.Printf("from host [%s], message [%s]\n", conn.RemoteAddr().String(), string(buf[:n]))
	if _, err := conn.Write([]byte("sent from discovery")); err != nil {
		panic(err)
	}
}
