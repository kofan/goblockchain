package gochain

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// Block represents a single block in a blockchain
type Block struct {
	Transactions []Transaction
	Timestamp    int64
	PrevHash     string
	Hash         string
	Difficulty   uint8
	Nonce        uint64
}

// NewBlock creates new block for the specified list of transactions
func NewBlock(transactions []*Transaction) *Block {
	ts := time.Now().Unix()
	trxs := make([]Transaction, len(transactions))

	for i, t := range transactions {
		trxs[i] = *t
	}

	return &Block{trxs, ts, "", "", 0, 0}
}

func (b *Block) after(previous *Block) {
	b.PrevHash = previous.Hash
}

func (b *Block) content() string {
	content := b.PrevHash + strconv.FormatInt(b.Timestamp, 10)
	for _, t := range b.Transactions {
		content += t.String()
	}

	content += strconv.FormatUint(uint64(b.Difficulty), 10)
	if b.Hash != "" {
		content += strconv.FormatUint(b.Nonce, 10)
	}
	return content
}

func (b *Block) mine(difficulty uint8) (time.Duration, error) {
	if difficulty < 0 || difficulty > 255 {
		return 0, fmt.Errorf("invalid difficulty value %d", difficulty)
	}

	var i uint8
	var nonce uint64
	var hash Blockhash

	b.Difficulty = difficulty
	content := b.content()
	start := time.Now()

mining:
	for {
		hash = NewBlockhash(content + strconv.FormatUint(nonce, 10))

		for i = 0; i < difficulty; i++ {
			if hash.isZeroBit(i) {
				continue
			}
			if nonce == math.MaxUint64 {
				return time.Since(start), fmt.Errorf("could not mine the block; \"math/big\" should be used")
			}
			nonce++
			continue mining
		}
		break
	}

	duration := time.Since(start)

	b.Nonce = nonce
	b.Hash = hash.toHex()

	return duration, nil
}

func (b *Block) verifyProofOfWork() bool {
	hash, err := NewBlockhashFromHexHash(b.Hash)
	if err != nil {
		return false
	}
	return hash.hasDifficulty(b.Difficulty)
}

func (b *Block) verifyHash() bool {
	if b.Hash == "" {
		return false
	}
	hash := NewBlockhash(b.content())
	return b.Hash == hash.toHex()
}
