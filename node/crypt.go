package node

import (
	"crypto/cipher"
	"io"
)

type encrypter struct {
	Next   io.Writer
	Cipher cipher.Stream
}

func (e *encrypter) Write(b []byte) (int, error) {
	buf := make([]byte, len(b))

	e.Cipher.XORKeyStream(buf, b)

	return e.Next.Write(buf)
}

type decrypter struct {
	Prev   io.Reader
	Cipher cipher.Stream
}

func (d *decrypter) Read(b []byte) (int, error) {
	n, err := d.Prev.Read(b)
	if err != nil {
		if n > 0 {
			d.Cipher.XORKeyStream(b, b[:n])
		}
		return n, err
	}

	d.Cipher.XORKeyStream(b, b[:n])

	return n, nil
}
