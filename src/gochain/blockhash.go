package gochain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Blockhash represents SHA256 hash
type Blockhash [sha256.Size]byte

// NewBlockhash creates new Blockhash instance for content of the string
func NewBlockhash(content string) Blockhash {
	return sha256.Sum256([]byte(content))
}

// NewBlockhashFromHexHash creates new Blockhash instance from the hex hash
func NewBlockhashFromHexHash(hash string) (Blockhash, error) {
	var blockhash [sha256.Size]byte
	bytes, err := hex.DecodeString(hash)

	if err != nil {
		return blockhash, errors.New("cannot decode hex hash")
	}

	copy(blockhash[:], bytes)
	return blockhash, nil
}

func (bh *Blockhash) isZeroBit(nbit uint16) bool {
	return (bh[nbit/8]>>uint(nbit%8))&1 == 0
}

func (bh *Blockhash) hasDifficulty(difficulty uint16) bool {
	for i := uint16(0); i < difficulty; i++ {
		if !bh.isZeroBit(i) {
			return false
		}
	}
	return true
}

func (bh *Blockhash) toHex() string {
	return fmt.Sprintf("%x", *bh)
}
