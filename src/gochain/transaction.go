package gochain

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Transaction is a struct blockchain transaction entity
type Transaction struct {
	UUID   string
	Target string
	Source string
	Amount uint64
}

// NewTransaction creates a new blockchain transaction
// specifying payload which goes from source to target
func NewTransaction(target, source string, amount uint64) *Transaction {
	trx := Transaction{
		UUID:   uuid.NewV1().String(),
		Target: target,
		Source: source,
		Amount: amount,
	}
	return &trx
}

func (t *Transaction) String() string {
	bytes, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
