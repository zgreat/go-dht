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

	var result node.Node

	_ = result

	go p.Gossip(gossipPeers)

	for {
		select {
		case peer := <-peers:
			if p.Peers == nil {
				log.Printf("p.Peers == nil")
			} else if peer == nil {
				log.Printf("peer == nil")
			} else if peer.GetID() == nil {
				log.Printf("peer.GetID() == nil")
			} else {
				if _, ok := p.Peers[peer.GetID().String()]; !ok {
					p.Peers[peer.GetID().String()] = peer
					log.Println("[+] new peer added:", peer.GetID())
				}
			}
		case <-heartbeat.C:
			for _, v := range p.Peers {
				if !v.IsValid() {
					log.Printf("[*] removing peer %s", v.GetID().String())
					delete(p.Peers, v.GetID().String())
					continue
				}

				log.Println("[*] sending heartbeat to", v.GetID().String())

				msg := node.NewMessage(p.ID, v.GetID(), "heartbeat", nil)
				v.SendMessage(msg, 1*time.Second)
			}
		case <-gossip.C:
			go func() {
				for _, v := range p.Peers {
					gossipPeers <- v
					time.Sleep(30 * time.Millisecond)
				}
			}()
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
				case "heartbeat":
					p.Peers[msg.GetSourceKey().String()].Refresh()

					log.Println("[*] received heartbeat from", msg.GetSourceKey().String())
				case "query_leader":
					if p.IsLeader {
						reply := node.NewMessage(p.ID, v.GetID(), "query_leader", make(map[string]interface{}))

						reply.GetArgs()["leader"] = p.ID

						v.SendMessage(reply, 1*time.Second)
					} else {
						if msg.GetArgs()["hops"].(uint64) > 3 {
							reply := node.NewMessage(p.ID, v.GetID(), "query_leader", nil)

							v.SendMessage(reply, 1*time.Second)
						} else {
							for _, others := range p.Peers {
								if v.GetID().Equals(others.GetID()) {
									continue
								}

								newMsg := node.NewMessage(p.ID, others.GetID(), "query_leader", nil)

								others.SendMessage(newMsg, 1*time.Second)

								reply, _ := others.ReceiveMessage(5 * time.Second)

								_ = reply

							}

						}
					}
				}
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
