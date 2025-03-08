package main

import (
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
)

type dataNode struct {
	net.Conn
	id uint16
}

func nodeInitialize(lbAddr string, nodeID uint16) (*dataNode, error) {
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

	dataNode := &dataNode{lbSoc, nodeID}

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
			go requestSim(dataNode)

		default:
			slog.Error("unsupported request.", "request", buf[0])

		}
	}
}

func requestSim(d *dataNode) {
	start := time.Now()

	bytesWritten, err := d.handleUserNodeJoin()
	if err != nil {
		slog.Error("failed to service user.", "err", err)
		return
	}

	dur := time.Since(start)
	// TODO: for Yasmin - add telemetry

	slog.Info("finished service user.", "file-size", bytesWritten, "duration", dur)
}

func (dataNode *dataNode) handleUserNodeJoin() (int, error) {
	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, ":0")
	if err != nil {
		return 0, err
	}

	defer soc.Close()

	// send port addr to LB
	var addrBuf [6]byte // 4 bytes for ipv4, 2 bytes for port

	addr, ok := soc.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("invalid TCP address")
	}

	if err := network.AddrToBytes(addr, addrBuf[:]); err != nil {
		return 0, err
	}

	if _, err := dataNode.Write(addrBuf[:]); err != nil {
		return 0, err
	}

	// handling user connection
	slog.Info("waiting for user connection.", "listener-addr", soc.Addr())

	user, err := soc.Accept()
	if err != nil {
		return 0, err
	}

	defer user.Close()

	slog.Info("user connected.", "addr", user.RemoteAddr())

	// get file digest + pub key from user
	var buf [64]byte
	if _, err := user.Read(buf[:]); err != nil {
		return 0, err
	}

	fileName := hex.EncodeToString(buf[:32])
	filePath := database.AccessCluster.Append(fileName).String()

	// read + decrypt file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	// send it over the wire
	plaintext := fileContent[crypto.NonceSize : len(fileContent)-crypto.OverHead]

	n, err := user.Write(plaintext)
	if err != nil {
		return 0, err
	}

	return n, nil
}
