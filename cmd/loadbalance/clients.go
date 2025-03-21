package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/hn275/distributed-storage/internal/network"
)

var cxMap = new(clientMap)

type clientMap struct{ sync.Map }

func (cx *clientMap) setClient(userConn net.Conn) {
	cx.Store(userConn.RemoteAddr().String(), userConn)
}

func (cx *clientMap) getClient(userAddr net.Addr) (net.Conn, bool) {
	v, ok := cx.LoadAndDelete(userAddr.String())
	return v.(net.Conn), ok
}

type dataNode struct {
	net.Conn
	rchan chan []byte
	wchan chan []byte
}

func makeDataNode(conn net.Conn) *dataNode {
	rchan := make(chan []byte, 100)
	wchan := make(chan []byte, 100)

	dataNode := &dataNode{conn, rchan, wchan}
	go dataNode.listen()

	// write routine
	go func(wchan <-chan []byte) {
		for {
			buf := <-wchan
			if _, err := conn.Write(buf); err != nil {
				// TODO: handle logging
				log.Println("makeDataNode error", err)
			}

		}
	}(wchan)

	return dataNode
}

func (dn *dataNode) write(buf []byte) {
	dn.wchan <- buf
}

func (dn *dataNode) listen() {
	for {
		buf := [16]byte{}
		_, err := dn.Read(buf[:])
		// TODO: handle logging
		if err != nil {
			log.Println(err)
			continue
		}

		switch buf[0] {
		case network.PortForwarding:
			go func() {
				if err := dn.handlePortForward(buf[:]); err != nil {
					log.Println(err)
				}
			}()

		default:
			log.Println("unsupported message type", buf[0], buf[1:])
		}
	}

}

func (dn *dataNode) handlePortForward(buf []byte) error {
	if len(buf) < 13 {
		panic("handlePortForward insufficient buf size")
	}

	clientAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		return err
	}

	client, ok := cxMap.getClient(clientAddr)
	if !ok {
		return fmt.Errorf("client not found in map [%s]", clientAddr)
	}

	defer client.Close()

	_, err = client.Write(buf[7:13])
	return err

}
