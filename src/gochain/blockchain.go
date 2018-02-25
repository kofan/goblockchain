package gochain

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// MaxDifficulty is the maximum value of the blockhain difficulty
const MaxDifficulty = 255

// CoinbaseSource is the source for coinbase transactions
const CoinbaseSource = "$coinbase$"

// GenesisBlockHash is a hash for the blockchain genesis block
// analogous to strings.Repeat("0", 64)
const GenesisBlockHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Blockchain represents blockchain structure
type Blockchain struct {
	stream              io.ReadWriteSeeker
	difficulty          uint8
	genesisBlock        *Block
	blocks              []Block
	pendingTransactions []Transaction
}

// NewBlockchain creates a new empty Blockchain instance
func NewBlockchain(stream io.ReadWriteSeeker, difficulty uint8) Blockchain {
	var bc Blockchain

	bc.blocks = make([]Block, 1, 128)
	bc.blocks[0] = Block{}

	bc.genesisBlock = &bc.blocks[0]
	bc.genesisBlock.Hash = GenesisBlockHash

	bc.pendingTransactions = make([]Transaction, 0, 32)

	bc.SetDifficulty(difficulty)
	if stream != nil {
		bc.attachStream(stream)
	}
	return bc
}

// SetDifficulty sets the blockchain difficulty to d
func (bc *Blockchain) SetDifficulty(d uint8) error {
	difficulty := uint8(d)
	if difficulty < bc.LastBlock().Difficulty {
		return fmt.Errorf("you cannot decrease the blockchain difficulty %d", bc.difficulty)
	}
	if difficulty > MaxDifficulty {
		return fmt.Errorf("the blockchain difficulty cannot be higher then %d", MaxDifficulty)
	}
	bc.difficulty = difficulty
	return nil
}

func (bc *Blockchain) attachStream(stream io.ReadWriteSeeker) {

}

// Verify checks that the block chain has a valid state
func (bc *Blockchain) Verify() bool {
	blocks := bc.blocks

	for i := 1; i < len(blocks); i++ {
		if blocks[i].PrevHash != blocks[i-1].Hash ||
			blocks[i].Difficulty < blocks[i-1].Difficulty ||
			blocks[i].Difficulty < bc.difficulty ||
			blocks[i].verifyProofOfWork() == false ||
			blocks[i].verifyHash() == false {
			return false
		}
	}

	return true
}

// PushTransaction creates new transaction remit the specified amount from target to source account
// if target account doesn't have enough money then false is returned
func (bc *Blockchain) PushTransaction(target, source string, amount uint64) bool {
	if bc.collectBalanceFor(target) < amount {
		return false
	}

	t := NewTransaction(target, source, amount)
	bc.pendingTransactions = append(bc.pendingTransactions, t)
	return true
}

// PushCoinbase adds amount to the target from nowhere i.e. coninbase transaction
func (bc *Blockchain) PushCoinbase(target string, amount uint64) {
	t := NewTransaction(target, CoinbaseSource, amount)
	bc.pendingTransactions = append(bc.pendingTransactions, t)
}

// LastBlock returns the last block of the keychain
func (bc *Blockchain) LastBlock() *Block {
	return &bc.blocks[len(bc.blocks)-1]
}

// IsEmpty checks whether there are any other blocks besides the genesis one
func (bc *Blockchain) IsEmpty() bool {
	return len(bc.blocks) == 1
}

func (bc *Blockchain) collectBalanceFor(target string) uint64 {
	targetBalance := uint64(0)

	for i := 1; i < len(bc.blocks); i++ {
		for j := 0; j < len(bc.blocks[i].Transactions); j++ {
			if bc.blocks[i].Transactions[j].Target == target {
				targetBalance += bc.blocks[i].Transactions[j].Amount
			} else if bc.blocks[i].Transactions[j].Source == target {
				targetBalance -= bc.blocks[i].Transactions[j].Amount
			}
		}
	}
	for j := 0; j < len(bc.pendingTransactions); j++ {
		if bc.pendingTransactions[j].Target == target {
			targetBalance += bc.pendingTransactions[j].Amount
		} else if bc.pendingTransactions[j].Source == target {
			targetBalance -= bc.pendingTransactions[j].Amount
		}
	}

	return targetBalance
}

// ProcessPendingTransactions mines a new block for all pending transactions and adds it to the blockchain
func (bc *Blockchain) ProcessPendingTransactions() (time.Duration, error) {
	if len(bc.pendingTransactions) == 0 {
		return 0, nil
	}

	block := NewBlock(bc.pendingTransactions)
	block.after(&bc.blocks[len(bc.blocks)-1])
	duration, err := block.mine(bc.difficulty)

	bc.blocks = append(bc.blocks, block)
	bc.pendingTransactions = bc.pendingTransactions[:0]

	return duration, err
}

func (bc *Blockchain) String() string {
	blocks, _ := json.MarshalIndent(bc.blocks, "", "  ")
	pendingTransactions, _ := json.MarshalIndent(bc.pendingTransactions, "", "  ")

	return strings.Join([]string{
		"\n",
		"--------------------------------------------------------",
		"======================== BLOCKS ========================",
		string(blocks),
		"================= PENDING TRANSACTIONS =================",
		string(pendingTransactions),
		"========================================================",
		"--------------------------------------------------------",
		"\n\n",
	}, "\n")
}
