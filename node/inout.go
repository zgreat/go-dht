package node

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"time"

	"github.com/ugorji/go/codec"
)

func (n *node) sendOut() {
	aesBlockCipher, _ := aes.NewCipher(n.GetCryptKey())

	crypt := &encrypter{
		Next:   n.conn,
		Cipher: cipher.NewCTR(aesBlockCipher, n.GetCryptIV()),
	}

	enc := codec.NewEncoder(crypt, &MessageFormat)

	for msg := range n.outbox {
		enc.Encode(msg)
	}
}

func (n *node) receiveIn() {

	aesBlockCipher, err := aes.NewCipher(n.GetCryptKey())
	if err != nil {
		log.Println("[!] error creating block cipher:", err)
	}

	crypt := &decrypter{
		Prev:   n.conn,
		Cipher: cipher.NewCTR(aesBlockCipher, n.GetCryptIV()),
	}

	dec := codec.NewDecoder(crypt, &MessageFormat)

	for {
		var msg NodeMessage
		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				n.Close()
				return
			}

			log.Println("[!] error receiving message:", err.Error())
			continue
		}

		n.lastSeen = time.Now()

		select {
		case <-n.done:
			return
		case n.inbox <- &msg:
		}
	}
}
