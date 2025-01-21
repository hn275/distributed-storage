package main

import (
	"log/slog"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "discovery:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	slog.Info("socket opened", "addr", conn.RemoteAddr())
	if _, err := conn.Write([]byte("sent from node")); err != nil {
		panic(err)
	}

	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	slog.Info("message received", "peer", conn.RemoteAddr(), "data", buf[:n])
}
