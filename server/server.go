package server

import (
	"crypto/aes"
	"crypto/rand"
	"io"
	"log"
	"net"

	"github.com/ugorji/go/codec"

	"github.com/gearmover/gofs2/node"
	"github.com/gearmover/gofs2/util"
)

// Server is the interface that handles adding new peers
type Server interface {
	Run(newPeers chan node.Node, systemPort int, cryptKey []byte)

	GetKey() util.Key

	Handshake(serverPort int, conn net.Conn) (util.Key, int)
}

type server struct {
	conn net.Listener
	key  util.Key

	iv       []byte
	cryptKey []byte
}

// New generates a new server object that accepts new peer connections
func New(laddr string, key util.Key) Server {
	conn, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Println("[!] unable to start server listening on", laddr, ":", err.Error())
		return nil
	}

	s := &server{
		conn: conn,
		key:  key,
	}

	s.iv = make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, s.iv)

	return s
}

func (s *server) GetKey() util.Key {
	return s.key
}

func (s *server) Run(newPeers chan node.Node, serverPort int, cryptKey []byte) {
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
			peerKey, systemPort := s.Handshake(serverPort, conn)
			if peerKey.Equals(util.NullKey) {
				log.Println("[*] error handshaking with new peer")
				return
			}

			log.Println("[+] new peer connected", peerKey)

			peer := node.NewNode(peerKey, uint16(systemPort), cryptKey, s.iv, conn)

			newPeers <- peer
		}()
	}
}

func (s *server) Handshake(serverPort int, conn net.Conn) (util.Key, int) {
	handshakeA := node.NewMessage(s.GetKey(), nil, "handshake", make(map[string]interface{}))

	handshakeA.GetArgs()["iv"] = s.iv
	handshakeA.GetArgs()["systemPort"] = serverPort

	log.Println("Server Key: ", s.GetKey())

	enc := codec.NewEncoder(conn, &node.MessageFormat)
	dec := codec.NewDecoder(conn, &node.MessageFormat)

	if err := enc.Encode(handshakeA); err != nil {
		log.Println("[*] handshake 1/3 failed:", err.Error())
		return util.NullKey, -1
	}

	var handshakeB node.NodeMessage

	if err := dec.Decode(&handshakeB); err != nil {
		log.Println("[*] handshake 2/3 failed:", err.Error())
		return util.NullKey, -1
	}

	clientKey := handshakeB.GetSourceKey()
	systemPort := handshakeB.GetArgs()["systemPort"].(uint64)

	handshakeC := node.NewMessage(s.GetKey(), handshakeB.GetSourceKey(), "ack", nil)

	if err := enc.Encode(handshakeC); err != nil {
		log.Println("[*] handshake 3/3 failed:", err.Error())
		return util.NullKey, -1
	}

	return clientKey, int(systemPort)
}
