package main

import (
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "discovery:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("sent from node")); err != nil {
		panic(err)
	}

	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	log.Printf("from host [%s], message [%s]\n", conn.RemoteAddr().String(), string(buf[:n]))
}
