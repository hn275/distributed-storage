package main

import (
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "discovery:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("Hello, World!")); err != nil {
		panic(err)
	}
}
