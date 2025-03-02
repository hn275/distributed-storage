package main

import (
	"encoding/hex"
	"log/slog"
	"net"
	"os"

	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
)

type dataNode struct {
	net.Conn
}

func nodeInitialize(lbAddr string) (*dataNode, error) {
	laddr, err := net.ResolveTCPAddr(network.ProtoTcp4, ":0") // randomize the port
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr(network.ProtoTcp4, lbAddr)
	if err != nil {
		return nil, err
	}

	// dial and ping LB, notifying node type
	lbSoc, err := net.DialTCP(network.ProtoTcp4, laddr, raddr)
	if err != nil {
		return nil, err
	}

	if _, err := lbSoc.Write([]byte{network.DataNodeJoin}); err != nil {
		return nil, err
	}

	dataNode := &dataNode{lbSoc}

	return dataNode, nil
}

func (dataNode *dataNode) Listen() {
	defer dataNode.Close()

	// 4 bytes for the address, 2 bytes for the port
	var buf [6]byte

	for {
		// get a request from LB
		_, err := dataNode.Read(buf[:])
		if err != nil {
			slog.Error("failed to read from LB.", "err", err)
			return
		}

		switch buf[0] {
		case network.UserNodeJoin:
			go dataNode.handleUserNodeJoin()

		default:
			slog.Error("unsupported request.", "request", buf[0])

		}
	}
}

func (dataNode *dataNode) handleUserNodeJoin() {
	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, ":0")
	if err != nil {
		slog.Error("failed to open new socket", "err", err)
		return
	}

	defer soc.Close()

	// send port number to LB
	var addrBuf [6]byte // 4 bytes for ipv4, 2 bytes for port

	addr, ok := soc.Addr().(*net.TCPAddr)
	if !ok {
		slog.Error("returned type is not net.TCPAddr")
		return
	}

	if err := network.AddrToBytes(addr, addrBuf[:]); err != nil {
		slog.Error(
			"error converting ip address (addr) to bytes.",
			"addr", addr,
			"err", err,
		)
		return
	}

	if _, err := dataNode.Write(addrBuf[:]); err != nil {
		slog.Error(
			"failed to write to LB.",
			"err", err,
		)
		return
	}

	// handling user connection
	slog.Info("waiting for user connection.",
		"addr", soc.Addr().String(),
		"protocol", soc.Addr().Network(),
	)

	user, err := soc.Accept()
	if err != nil {
		panic(err)
	}

	slog.Info("user connected.",
		"addr", user.RemoteAddr(),
		"protocol", user.RemoteAddr().Network(),
	)

	defer user.Close()

	// get file digest + pub key from user
	var buf [64]byte
	if _, err := user.Read(buf[:]); err != nil {
		panic(err)
	}

	fileName := hex.EncodeToString(buf[:32])
	filePath := database.AccessCluster.Append(fileName).String()

	// read + decrypt file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var (
		pubKey     []byte = buf[32:]
		secKey     []byte = crypto.DataNodeSecretKey[:]
		dst        []byte = fileContent[crypto.NonceSize:crypto.NonceSize]
		nonce      []byte = fileContent[:crypto.NonceSize]
		ciphertext []byte = fileContent[crypto.NonceSize:]
	)

	err = crypto.Decrypt(
		dst, secKey, nonce,
		ciphertext, pubKey,
	)

	if err != nil {
		panic(err)
	}

	// send it over the wire
	plaintext := fileContent[crypto.NonceSize : len(fileContent)-crypto.OverHead]

	if _, err := user.Write(plaintext); err != nil {
		panic(err)
	}

	slog.Info("user service complete.", "addr", user.RemoteAddr())
}
