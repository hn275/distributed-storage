package main

import (
	"fmt"
	"net"

	"github.com/hn275/dist-db/internal"
)

func main() {
	soc, err := net.Listen("tcp", ":"+internal.MustEnv("SOCKET_PORT"))
	if err != nil {
		panic(err)
	}
	defer soc.Close()

	fmt.Println("Local socket address", soc.Addr())
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

	fmt.Println("Connection established:", conn.LocalAddr())
}
