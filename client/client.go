package client

import (
	"log"
	"net"

	"github.com/gearmover/gofs2/node"
	"github.com/gearmover/gofs2/util"
	"github.com/ugorji/go/codec"
)

// New connects to an already existing node and handshakes with it,
// creating our first connection
func New(host string, key util.Key, cryptKey []byte) node.Node {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("[!] unable to dial host", host, ":", err.Error())
		return nil
	}

	enc := codec.NewEncoder(conn, &node.MessageFormat)
	dec := codec.NewDecoder(conn, &node.MessageFormat)

	var handshakeA node.NodeMessage

	// do handshake
	if err := dec.Decode(&handshakeA); err != nil {
		log.Println("[!] client handshake 1/3 failed:", err.Error())
		return nil
	}

	iv := handshakeA.GetArgs()["iv"].([]byte)

	serverKey := handshakeA.GetSourceKey()
	handshakeB := node.NewMessage(key, serverKey, "handshake", nil)

	if err := enc.Encode(handshakeB); err != nil {
		log.Println("[!] client handshake 2/3 failed:", err.Error())
		return nil
	}

	var handshakeC node.NodeMessage

	if err := dec.Decode(&handshakeC); err != nil {
		log.Println("[!] client handshake 3/3 failed:", err.Error())
		return nil
	}

	if handshakeC.GetCommand() != "ack" {
		log.Println("[!] invalid handshake message received:", handshakeC.GetCommand())
		return nil
	}

	n := node.NewNode(serverKey, cryptKey, iv, conn)

	return n
}
