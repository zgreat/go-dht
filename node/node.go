package node

import (
	"errors"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gearmover/gofs2/util"
)

// Node represents a node on the network
type Node interface {
	GetID() util.Key
	GetClosestPeers(k uint) []Node

	GetCryptKey() []byte
	GetCryptIV() []byte

	GetSystemPort() uint16

	AddPeer(n Node) error
	GetPeer(id util.Key) Node
	RmPeer(id util.Key) error

	GetHostPort() (net.IP, uint16)

	IsValid() bool
	Close() error

	DistanceFrom(key util.Key) util.Key

	Refresh()

	SendMessage(m Message, timeout time.Duration) error
	ReceiveMessage(timeout time.Duration) (Message, error)
	RequeueMessage(m Message, timeout time.Duration) error
}

type node struct {
	id    util.Key
	peers []Node
	conn  net.Conn

	valid bool

	key []byte
	iv  []byte

	lastSeen time.Time

	systemPort uint16

	outbox chan Message
	inbox  chan Message
	done   chan struct{}
}

const (
	// NodeTimeout the timeout time for a node that isn't responsive
	NodeTimeout = 15 * time.Second
)

var (
	// ErrPeerAlreadyKnown .
	ErrPeerAlreadyKnown = errors.New("peer already known")
	// ErrPeerNotFound .
	ErrPeerNotFound = errors.New("peer not found")
	// ErrTimeout .
	ErrTimeout = errors.New("request timed out")
)

// NewNode generates a new node object using the supplied connection details
func NewNode(id util.Key, systemPort uint16, key []byte, iv []byte, conn net.Conn) Node {
	n := &node{
		id:         id,
		peers:      make([]Node, 0),
		conn:       conn,
		key:        key,
		iv:         iv,
		lastSeen:   time.Now(),
		systemPort: systemPort,
		outbox:     make(chan Message, 1),
		inbox:      make(chan Message, 1),
		done:       make(chan struct{}),
		valid:      true,
	}

	go n.receiveIn()
	go n.sendOut()

	return n
}

func (n *node) GetSystemPort() uint16 {
	return n.systemPort
}

func (n *node) GetID() util.Key {
	return n.id
}

func (n *node) IsValid() bool {
	if time.Now().Sub(n.lastSeen) > NodeTimeout {
		return false
	}

	return n.valid
}

func (n *node) Refresh() {
	n.lastSeen = time.Now()
}

func (n *node) GetCryptKey() []byte {
	return n.key
}

func (n *node) GetCryptIV() []byte {
	return n.iv
}

func (n *node) GetHostPort() (net.IP, uint16) {
	parts := strings.Split(n.conn.RemoteAddr().String(), ":")

	ip := net.ParseIP(parts[0])
	port, _ := strconv.Atoi(parts[1])

	return ip, uint16(port)
}

func (n *node) GetClosestPeers(k uint) []Node {
	ids := make([]util.Key, len(n.peers))

	for i, y := range n.peers {
		ids[i] = y.DistanceFrom(n.GetID())
	}

	sort.Sort(util.ByKey(ids))

	var nodes []Node

	max := int(k)

	if max > len(ids) {
		max = len(ids)
	}

	for i := 0; i < max; i++ {
		nodes = append(nodes, n.GetPeer(ids[i]))
	}

	return nodes
}

func (n *node) AddPeer(peer Node) error {
	if t := n.GetPeer(peer.GetID()); t != nil {
		return ErrPeerAlreadyKnown
	}

	n.peers = append(n.peers, peer)

	return nil
}

func (n *node) GetPeer(id util.Key) Node {
	for _, k := range n.peers {
		if k.GetID().Equals(id) {
			return k
		}
	}

	return nil
}

func (n *node) RmPeer(id util.Key) error {
	for i, v := range n.peers {
		if v.GetID().Equals(id) {
			n.peers = append(n.peers[:i], n.peers[(i+1):]...)

			return nil
		}
	}

	return ErrPeerNotFound
}

func (n *node) DistanceFrom(k util.Key) util.Key {
	return n.GetID().Xor(k)
}

func (n *node) SendMessage(m Message, timeout time.Duration) error {
	if !n.IsValid() {
		return ErrPeerNotFound
	}

	timer := time.NewTimer(timeout)

	select {
	case n.outbox <- m:
	case <-timer.C:
		return ErrTimeout
	}

	return nil
}

func (n *node) ReceiveMessage(timeout time.Duration) (Message, error) {
	timer := time.NewTimer(timeout)

	var msg Message

	select {
	case msg = <-n.inbox:
	case <-timer.C:
		return nil, ErrTimeout
	}

	return msg, nil
}

func (n *node) RequeueMessage(m Message, timeout time.Duration) error {
	timer := time.NewTimer(timeout)

	select {
	case n.inbox <- m:
	case <-timer.C:
		return ErrTimeout
	}

	return nil
}

func (n *node) Close() error {
	close(n.done)
	close(n.outbox)

	n.valid = false

	return nil
}
