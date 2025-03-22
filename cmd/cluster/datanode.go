package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
)

var nodeJoinSignal = [1]byte{network.DataNodeJoin}

const randomLocalPort = "127.0.0.1:0"

type dataNode struct {
	net.Conn
	id    uint16
	wchan chan []byte
}

func makeDataNode(c net.Conn, id uint16) *dataNode {
	dataNode := &dataNode{c, id, make(chan []byte, 100)}

	go func(wchan <-chan []byte) {
		for {
			buf := <-wchan
			if _, err := c.Write(buf); err != nil {
				log.Println(err) // TODO: handle logging
			}
		}
	}(dataNode.wchan)

	return dataNode
}

func (d *dataNode) write(buf []byte) {
	d.wchan <- buf
}

func nodeInitialize(lbAddr string, nodeID uint16) (*dataNode, error) {
	laddr, err := net.ResolveTCPAddr(network.ProtoTcp4, randomLocalPort)
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

	dataNode := makeDataNode(lbSoc, nodeID)

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
			go func() {
				if err := dataNode.handleUserJoin(buf[:]); err != nil {
					log.Println(err) // TODO: log error
				}
			}()
			// go requestSim(dataNode)

		default:
			slog.Error("unsupported request.", "request", buf[0])

		}
	}
}

func (d *dataNode) handleUserJoin(buf []byte) error {
	if len(buf) < 13 {
		panic("handleUserJoin insufficient buf size")
	}

	userAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		return err
	}

	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, randomLocalPort)
	if err != nil {
		return err
	}

	// PORT FORWARD TO LB
	buf[0] = network.PortForwarding
	if err := network.AddrToBytes(soc.Addr(), buf[7:13]); err != nil {
		return err
	}

	d.write(buf)

	// SERVING CLIENT
	go func(soc net.Listener, userAddr net.Addr) {
		if err := d.serveClient(soc, userAddr); err != nil {
			panic(err)
		}
	}(soc, userAddr)

	return nil
}

func (d *dataNode) serveClient(soc net.Listener, userAddr net.Addr) error {
	defer soc.Close()

	user, err := soc.Accept()
	if err != nil {
		return err
	}

	defer user.Close()

	// check for the connection, need to match with `userAddr`
	sameUser := user.
		RemoteAddr().(*net.TCPAddr).
		IP.
		Equal(userAddr.(*net.TCPAddr).IP)
	if !sameUser {
		return fmt.Errorf(
			"invalid user ip address, expected [%v], got [%v]",
			userAddr.String(), user.RemoteAddr().String())
	}

	slog.Info("user connected.", "addr", user.RemoteAddr())

	// get file digest + pub key from user
	var fileBuf [64]byte
	if _, err := user.Read(fileBuf[:]); err != nil {
		return err
	}

	fileName := hex.EncodeToString(fileBuf[:32])
	filePath := database.AccessCluster.Append(fileName).String()

	// read + decrypt file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var (
		pubKey     []byte = fileBuf[32:]
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
		return err
	}

	// send it over the wire
	plaintext := fileContent[crypto.NonceSize : len(fileContent)-crypto.OverHead]

	_, err = user.Write(plaintext)
	if err != nil {
		return err
	}

	return nil
}
