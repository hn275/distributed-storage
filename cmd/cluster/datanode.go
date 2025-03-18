package main

import (
	"encoding/hex"
	"errors"
	"io"
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

var nodeJoinSignal = [1]byte{network.DataNodeJoin}

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

	if _, err := lbSoc.Write(nodeJoinSignal[:]); err != nil {
		return nil, err
	}

	dataNode := &dataNode{lbSoc, nodeID}

	return dataNode, nil
}

func (dataNode *dataNode) Listen() {
	defer dataNode.Close()

	for {
		var buf [16]byte
		// get a request from LB
		_, err := dataNode.Read(buf[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info(
					"load balancer disconnected.",
					"node-id", dataNode.id,
				)
			} else {
				slog.Error("failed to read from LB.", "err", err)
			}
			return
		}

		switch buf[0] {
		case network.UserNodeJoin:
			go requestSim(dataNode, buf[:])

		default:
			slog.Error("unsupported request.", "request", buf[0])

		}
	}
}

func requestSim(d *dataNode, msgBuf []byte) {
	start := time.Now()

	bytesWritten, err := d.handleUserNodeJoin(msgBuf)
	if err != nil {
		slog.Error("failed to service user.", "err", err)
		return
	}

	dur := time.Since(start)
	// TODO: for Yasmin - add telemetry

	slog.Info("finished service user.", "file-size", bytesWritten, "duration", dur)
}

func (dataNode *dataNode) handleUserNodeJoin(msgBuf []byte) (int, error) {
	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, ":0")
	if err != nil {
		return 0, err
	}

	defer soc.Close()

	// write port to msgBuf
	addr, ok := soc.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("invalid TCP address")
	}

	if err := network.AddrToBytes(addr, msgBuf[7:13]); err != nil {
		return 0, err
	}

	msgBuf[0] = network.DataNodePort
	if _, err := dataNode.Write(msgBuf[:]); err != nil {
		return 0, err
	}

	// handling user connection
	userAddr, err := network.BytesToAddr(msgBuf[1:7])
	if err != nil {
		panic(err)
	}
	slog.Info(
		"waiting for user connection.",
		"user", userAddr,
		"listener-addr", soc.Addr(),
	)

	user, err := soc.Accept()
	if err != nil {
		return 0, err
	}

	defer user.Close()

	// TODO: verify user address
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
