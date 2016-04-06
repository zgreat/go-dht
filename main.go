package main

import (
	"log"
	"os"

	"github.com/gearmover/gofs2/client"
	"github.com/gearmover/gofs2/node"
	"github.com/gearmover/gofs2/protocol"
	"github.com/gearmover/gofs2/server"
	"github.com/gearmover/gofs2/util"
)

func main() {

	laddr := "127.0.0.1:8080"

	if len(os.Args) > 1 {
		laddr = os.Args[1]
	}

	ourKey := util.NewKey()

	log.Println("Our Key: ", ourKey.String())

	server := server.New(laddr, ourKey)

	peers := make(chan node.Node, 10)

	go server.Run(peers, []byte("hello world 1234"))

	if len(os.Args) > 2 {
		serverAddr := os.Args[2]

		log.Println("[+] attempting to bootstrap with server", serverAddr)

		peer := client.New(serverAddr, ourKey, []byte("hello world 1234"))
		if peer == nil {
			log.Println("[!] unable to bootstrap with server", serverAddr)
			return
		}

		log.Println("[+] bootstrap successful.")

		peers <- peer
	}

	proto := protocol.New(ourKey)

	proto.Run(peers)
}
