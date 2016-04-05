package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
)

// Server is the interface that handles adding new peers
type Server interface {
	Run(newPeers chan Node, msgStream chan Message)

	GetKey() Key

	Handshake(conn net.Conn) Key
}

type server struct {
	conn net.Listener
	key  Key
}

// NewServer generates a new server object that accepts new peer connections
func NewServer(laddr string, key Key) Server {
	conn, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Println("[!] unable to start server listening on", laddr, ":", err.Error())
		return nil
	}

	s := &server{
		conn: conn,
		key:  key,
	}

	return s
}

func (s *server) GetKey() Key {
	return s.key
}

func (s *server) Run(newPeers chan Node, msgStream chan Message) {
	for {
		conn, err := s.conn.Accept()
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Println("[*] error accepting new peer:", err.Error())
			continue
		}

		func() {
			peerKey := s.Handshake(conn)
			if peerKey.Equals(NullKey) {
				log.Println("[*] error handshaking with new peer")
				return
			}

			log.Println("[+] new peer connected", peerKey)

			peer := NewNode(peerKey, conn, msgStream)

			newPeers <- peer
		}()
	}
}

func (s *server) Handshake(conn net.Conn) Key {
	handshakeA := NewMessage(s.GetKey(), nil, "handshake", nil)

	log.Println("Server Key: ", s.GetKey())

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(handshakeA); err != nil {
		log.Println("[*] handshake 1/3 failed:", err.Error())
		return NullKey
	}

	var handshakeB nodeMessage

	if err := dec.Decode(&handshakeB); err != nil {
		log.Println("[*] handshake 2/3 failed:", err.Error())
		return NullKey
	}

	clientKey := handshakeB.GetSourceKey()

	handshakeC := NewMessage(s.GetKey(), handshakeB.GetSourceKey(), "ack", nil)

	if err := enc.Encode(handshakeC); err != nil {
		log.Println("[*] handshake 3/3 failed:", err.Error())
		return NullKey
	}

	return clientKey
}
