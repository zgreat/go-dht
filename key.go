package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"
)

// Key represents a unique identifier for an object or node
type Key []byte

var (
	// NullKey is used when we don't know or care about a key
	NullKey Key
)

// Equals returns true if two keys are byte-equivalent, false otherwise
func (u Key) Equals(k Key) bool { return bytes.Compare(u, k) == 0 }

// Xor returns the xor of all bytes of key A with key B
func (u Key) Xor(k Key) Key {
	xor := make([]byte, len(k))

	for i, v := range u {
		xor[i] = v ^ k[i]
	}

	return Key(xor)
}

func (u Key) String() string {
	return hex.EncodeToString(u)
}

var keyRand rand.Source

// NewKey generates a new unique key
func NewKey() Key {
	// check if we need to initialize our random number generator
	if keyRand == nil {
		keyRand = rand.NewSource(int64(time.Now().Unix()))
	}

	// create a new MD5 object
	m := md5.New()

	// write our random integer into the hash buffer
	binary.Write(m, binary.LittleEndian, keyRand.Int63())

	// return the generated MD5 sum
	return m.Sum(nil)
}

// ByKey allows sorting of Keys using the sort package
type ByKey []Key

func (u ByKey) Len() int           { return len(u) }
func (u ByKey) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u ByKey) Less(i, j int) bool { return bytes.Compare(u[i], u[j]) < 0 }
