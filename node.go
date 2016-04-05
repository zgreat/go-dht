package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sort"
	"time"
)

// Node represents a node on the network
type Node interface {
	GetID() Key
	GetClosestPeers(k uint) []Node

	AddPeer(n Node) error
	GetPeer(id Key) Node
	RmPeer(id Key) error

	IsValid() bool
	Close() error

	DistanceFrom(key Key) Key

	SendMessage(m Message, timeout time.Duration) error
}

type node struct {
	id    Key
	peers []Node
	conn  net.Conn

	valid bool

	outbox chan Message
	inbox  chan Message
	done   chan struct{}
}

var (
	// ErrPeerAlreadyKnown .
	ErrPeerAlreadyKnown = errors.New("peer already known")
	// ErrPeerNotFound .
	ErrPeerNotFound = errors.New("peer not found")
	// ErrTimeout .
	ErrTimeout = errors.New("request timed out")
)

// NewNode generates a new node object using the supplied connection details
func NewNode(id Key, conn net.Conn, inbound chan Message) Node {
	n := &node{
		id:     id,
		peers:  make([]Node, 0),
		conn:   conn,
		outbox: make(chan Message, 1),
		inbox:  inbound,
		done:   make(chan struct{}),
		valid:  true,
	}

	go n.receiveIn()
	go n.sendOut()

	return n
}

func (n *node) GetID() Key {
	return n.id
}

func (n *node) IsValid() bool {
	return n.valid
}

func (n *node) GetClosestPeers(k uint) []Node {
	ids := make([]Key, len(n.peers))

	for i, k := range n.peers {
		ids[i] = k.DistanceFrom(n.GetID())
	}

	sort.Sort(ByKey(ids))

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

func (n *node) GetPeer(id Key) Node {
	for _, k := range n.peers {
		if k.GetID().Equals(id) {
			return k
		}
	}

	return nil
}

func (n *node) RmPeer(id Key) error {
	for i, v := range n.peers {
		if v.GetID().Equals(id) {
			n.peers = append(n.peers[:i], n.peers[(i+1):]...)

			return nil
		}
	}

	return ErrPeerNotFound
}

func (n *node) DistanceFrom(k Key) Key {
	return n.GetID().Xor(k)
}

func (n *node) SendMessage(m Message, timeout time.Duration) error {
	timer := time.NewTimer(timeout)

	select {
	case n.outbox <- m:
	case <-timer.C:
		return ErrTimeout
	}

	return nil
}

func (n *node) sendOut() {
	enc := json.NewEncoder(n.conn)

	for msg := range n.outbox {
		enc.Encode(msg)
	}
}

func (n *node) receiveIn() {
	dec := json.NewDecoder(n.conn)

	for {
		var msg nodeMessage
		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				n.Close()
				return
			}

			log.Println("[!] error receiving message:", err.Error())
			continue
		}

		select {
		case <-n.done:
			return
		case n.inbox <- &msg:
		}
	}
}

func (n *node) Close() error {
	close(n.done)
	close(n.outbox)

	n.valid = false

	return nil
}
