package main

import (
	"log/slog"
	"net"

	"github.com/hn275/dist-db/internal"
)

func main() {
	soc, err := net.Listen("tcp", ":"+internal.MustEnv("SOCKET_PORT"))
	if err != nil {
		panic(err)
	}
	defer soc.Close()

	slog.Info("local socket address", "addr", soc.Addr())
	for {
		conn, err := soc.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			return
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	slog.Info("connection established", "peer", conn.RemoteAddr())
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	if _, err := conn.Write([]byte("sent from discovery")); err != nil {
		panic(err)
	}
	slog.Info("message sent", "peer", conn.RemoteAddr(), "data", string(buf[:n]))
	slog.Info("connection closed", "peer", conn.RemoteAddr())
}
