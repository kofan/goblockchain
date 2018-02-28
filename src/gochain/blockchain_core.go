package gochain

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"time"
)

// MaxDifficulty is the maximum value of the blockhain difficulty
const MaxDifficulty = 255

// CoinbaseSource is the source for coinbase transactions
const CoinbaseSource = "$coinbase$"

// GenesisBlockHash is a hash for the blockchain genesis block
// analogous to strings.Repeat("0", 64)
const GenesisBlockHash = "0000000000000000000000000000000000000000000000000000000000000000"

var tplFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}
var tpl = template.Must(template.
	New("templates").
	Funcs(tplFuncs).
	ParseGlob("templates/gochain/*.gotmpl"),
)

// Blockchain represents blockchain structure
type Blockchain struct {
	stream io.ReadWriteSeeker

	currentNode   Node
	adjacentNodes []Node

	difficulty   uint8
	genesisBlock *Block
	blocks       []*Block
	pending      []*Transaction
}

// NewBlockchain creates a new empty Blockchain instance
func NewBlockchain(node Node, difficulty uint8) *Blockchain {
	var bc Blockchain

	bc.currentNode = node
	bc.blocks = make([]*Block, 1, 128)
	bc.blocks[0] = &Block{}

	bc.genesisBlock = bc.blocks[0]
	bc.genesisBlock.Hash = GenesisBlockHash

	bc.pending = make([]*Transaction, 0, 32)

	bc.SetDifficulty(difficulty)
	return &bc
}

// NewBlockchainFromStream creates a new Blockchin instance from a stream of data
// trying to read the the serialized representation of the blockchain
func NewBlockchainFromStream(stream io.ReadWriteSeeker) {

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

// PushTransaction creates new transaction remit the specified amount from source to target account
// if target account doesn't have enough money then false is returned
func (bc *Blockchain) PushTransaction(target, source string, amount uint64) bool {
	if bc.ComputeBalanceFor(source) < amount {
		return false
	}
	bc.appendTransaction(NewTransaction(target, source, amount))
	return true
}

// PushCoinbase adds amount to the target from nowhere i.e. coninbase transaction
func (bc *Blockchain) PushCoinbase(target string, amount uint64) {
	bc.appendTransaction(NewTransaction(target, CoinbaseSource, amount))
}

func (bc *Blockchain) appendTransaction(t *Transaction) {
	bc.pending = append(bc.pending, t)
	bc.SyncWithAdjacentNodes(SyncModePushFirst)
}

// LastBlock returns the last block of the keychain
func (bc *Blockchain) LastBlock() *Block {
	return bc.blocks[len(bc.blocks)-1]
}

// IsEmpty checks whether there are any other blocks besides the genesis one
func (bc *Blockchain) IsEmpty() bool {
	return len(bc.blocks) == 1
}

// ComputeBalanceFor computes the current amount for specified source
func (bc *Blockchain) ComputeBalanceFor(source string) uint64 {
	sourceBalance := uint64(0)

	for i := 1; i < len(bc.blocks); i++ {
		for j := 0; j < len(bc.blocks[i].Transactions); j++ {
			if bc.blocks[i].Transactions[j].Target == source {
				sourceBalance += bc.blocks[i].Transactions[j].Amount
			} else if bc.blocks[i].Transactions[j].Source == source {
				sourceBalance -= bc.blocks[i].Transactions[j].Amount
			}
		}
	}
	for j := 0; j < len(bc.pending); j++ {
		if bc.pending[j].Target == source {
			sourceBalance += bc.pending[j].Amount
		} else if bc.pending[j].Source == source {
			sourceBalance -= bc.pending[j].Amount
		}
	}

	return sourceBalance
}

// ProcessPendingTrasactions mines a new block for all pending transactions and adds it to the blockchain
func (bc *Blockchain) ProcessPendingTrasactions() (time.Duration, error) {
	if len(bc.pending) == 0 {
		return 0, nil
	}

	block := NewBlock(bc.pending)
	block.after(bc.blocks[len(bc.blocks)-1])
	duration, err := block.mine(bc.difficulty)

	bc.blocks = append(bc.blocks, block)
	bc.pending = bc.pending[:0]
	bc.SyncWithAdjacentNodes(SyncModePushFirst)

	return duration, err
}

func (bc *Blockchain) FormatConsole() string {
	var buf bytes.Buffer

	err := tpl.ExecuteTemplate(&buf, "console.gotmpl", struct {
		CoinbaseSource string
		Blocks         []*Block
		Pending        []*Transaction
	}{CoinbaseSource, bc.blocks, bc.pending})

	if err != nil {
		panic(err)
	}

	return buf.String()
}
