package main

import (
	"log"
	"net"

	"encoding/json"
)

// Bootstrap connects to an already existing node and handshakes with it,
// creating our first connection
func Bootstrap(host string, key Key, msgStream chan Message) Node {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("[!] unable to dial host", host, ":", err.Error())
		return nil
	}

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	var handshakeA nodeMessage

	// do handshake
	if err := dec.Decode(&handshakeA); err != nil {
		log.Println("[!] client handshake 1/3 failed:", err.Error())
		return nil
	}

	serverKey := handshakeA.GetSourceKey()

	log.Println(handshakeA)

	handshakeB := NewMessage(key, serverKey, "handshake", nil)

	if err := enc.Encode(handshakeB); err != nil {
		log.Println("[!] client handshake 2/3 failed:", err.Error())
		return nil
	}

	var handshakeC nodeMessage

	if err := dec.Decode(&handshakeC); err != nil {
		log.Println("[!] client handshake 3/3 failed:", err.Error())
		return nil
	}

	if handshakeC.GetCommand() != "ack" {
		log.Println("[!] invalid handshake message received:", handshakeC.GetCommand())
		return nil
	}

	n := NewNode(serverKey, conn, msgStream)

	return n
}
