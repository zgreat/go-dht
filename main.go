package main

import (
	"log"
	"os"
	"time"
)

func main() {

	laddr := "127.0.0.1:8080"

	if len(os.Args) > 1 {
		laddr = os.Args[1]
	}

	ourKey := NewKey()

	log.Println("Our Key: ", ourKey.String())

	server := NewServer(laddr, ourKey)

	peers := make(chan Node, 10)
	msgs := make(chan Message, 10)

	go server.Run(peers, msgs)

	if len(os.Args) > 2 {
		serverAddr := os.Args[2]

		log.Println("[+] attempting to bootstrap with server", serverAddr)

		peer := Bootstrap(serverAddr, ourKey, msgs)
		if peer == nil {
			log.Println("[!] unable to bootstrap with server", serverAddr)
			return
		}

		log.Println("[+] bootstrap successful.")

		peers <- peer
	}

	Nodes := make([]Node, 0)

	heartbeat := time.NewTicker(5 * time.Second)

	for {
		select {
		case m := <-msgs:
			log.Println(" Message:", m)
		case p := <-peers:
			log.Println("New Peer:", p.GetID().String())

			Nodes = append(Nodes, p)
		case <-heartbeat.C:
			log.Println(Nodes)
			for i, v := range Nodes {
				if !v.IsValid() {
					log.Println("[*] deleting invalid node from node list:", v.GetID().String())
					Nodes = append(Nodes[:i], Nodes[(i+1):]...)
					continue
				}
				log.Println("[+] sending heartbeat to", v.GetID().String())
				beat := NewMessage(ourKey, v.GetID(), "heartbeat", nil)
				v.SendMessage(beat, time.Second*1)
			}
		}
	}
}
