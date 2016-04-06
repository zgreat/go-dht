package protocol

import (
	"log"
	"net"
	"time"

	"github.com/gearmover/gofs2/node"
	"github.com/gearmover/gofs2/util"
)

type Protocol interface {
	Run(peers chan node.Node)
	Gossip(peers chan node.Node)
}

type protocol struct {
	Peers    map[string]node.Node
	ID       util.Key
	IsLeader bool
}

type gossipNode struct {
	ID     util.Key
	IPAddr net.IP
	Port   uint16
}

func New(id util.Key) Protocol {
	return &protocol{
		Peers:    make(map[string]node.Node),
		ID:       id,
		IsLeader: false,
	}
}

func (p *protocol) Run(peers chan node.Node) {
	heartbeat := time.NewTicker(5 * time.Second)
	gossip := time.NewTicker(10 * time.Second)

	gossipPeers := make(chan node.Node, 1)

	go p.Gossip(gossipPeers)

	for {
		select {
		case peer := <-peers:
			p.Peers[peer.GetID().String()] = peer

			log.Println("[+] new peer added:", peer.GetID())
		case <-heartbeat.C:
			for _, v := range p.Peers {
				if !v.IsValid() {
					delete(p.Peers, v.GetID().String())
					continue
				}

				msg := node.NewMessage(p.ID, v.GetID(), "heartbeat", nil)
				v.SendMessage(msg, 1*time.Second)
			}
		case <-gossip.C:
			for _, v := range p.Peers {
				gossipPeers <- v
			}
		default:
			for _, v := range p.Peers {
				msg, err := v.ReceiveMessage(1 * time.Millisecond)
				if err == node.ErrTimeout {
					continue
				}

				log.Println("Received message from", v.GetID(), ":", msg.GetCommand())
			}
		}
	}
}

func (p *protocol) Gossip(peers chan node.Node) {
	for peer := range peers {
		msg := node.NewMessage(p.ID, peer.GetID(), "gossip", make(map[string]interface{}))
		nodes := make([]gossipNode, 0, len(p.Peers))

		for _, v := range p.Peers {

			if !v.IsValid() {
				continue
			}

			if v.GetID().Equals(peer.GetID()) {
				continue
			}

			ip, port := v.GetHostPort()

			gossip := gossipNode{
				ID:     v.GetID(),
				IPAddr: ip,
				Port:   port,
			}

			nodes = append(nodes, gossip)
		}

		msg.GetArgs()["gossip"] = nodes

		peer.SendMessage(msg, 1*time.Second)
	}
}
