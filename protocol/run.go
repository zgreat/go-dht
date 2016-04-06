package protocol

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gearmover/gofs2/client"
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
	CryptKey []byte
	IsLeader bool
}

type gossipNode struct {
	ID     util.Key
	IPAddr net.IP
	Port   uint16
}

func New(id util.Key, cryptKey []byte) Protocol {
	return &protocol{
		Peers:    make(map[string]node.Node),
		ID:       id,
		IsLeader: false,
		CryptKey: cryptKey,
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

				switch msg.GetCommand() {
				case "gossip":
					newPeers := msg.GetArgs()["gossip"].([]interface{})

					for _, np := range newPeers {

						tnp := np.([]interface{})

						inp := gossipNode{
							ID:     util.Key(tnp[0].([]uint8)),
							IPAddr: net.IP(tnp[1].([]uint8)),
							Port:   uint16(tnp[2].(uint64)),
						}
						if _, ok := p.Peers[inp.ID.String()]; !ok {
							peers <- client.New(fmt.Sprintf("%s:%d", inp.IPAddr.String(), inp.Port), inp.Port, p.ID, p.CryptKey)
						}
					}
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

			ip, _ := v.GetHostPort()

			gossip := gossipNode{
				ID:     v.GetID(),
				IPAddr: ip,
				Port:   v.GetSystemPort(),
			}

			nodes = append(nodes, gossip)
		}

		msg.GetArgs()["gossip"] = nodes

		peer.SendMessage(msg, 1*time.Second)
	}
}
