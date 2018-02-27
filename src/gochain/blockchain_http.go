package gochain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/kofan/goblockchain/src/common/httputil"
)

var errChainOutdated = errors.New("node's chain is outdated")
var muChainAccept sync.Mutex

type syncMode int

const (
	SyncModePushFirst syncMode = iota
	SyncModePullFirst
)

type blockchainPayload struct {
	Difficulty uint8         `json:"difficulty"`
	Blocks     []Block       `json:"blocks"`
	Pending    []Transaction `json:"transactions"`
}

func (bcp *blockchainPayload) marshal() ([]byte, error) {
	return json.Marshal(bcp)
}
func (bcp *blockchainPayload) unmarshal(data []byte) error {
	if err := json.Unmarshal(data, bcp); err != nil {
		return fmt.Errorf("cannot parse blockchain data: %v", err)
	}
	return nil
}

func (bc *Blockchain) setPayload(bcp *blockchainPayload) {
	bc.difficulty = bcp.Difficulty
	bc.blocks = bc.blocks[:0]
	bc.pending = bc.pending[:0]

	for _, b := range bcp.Blocks {
		bc.blocks = append(bc.blocks, &b)
	}
	for _, t := range bcp.Pending {
		bc.pending = append(bc.pending, &t)
	}
}
func (bc *Blockchain) getPayload() *blockchainPayload {
	bcp := &blockchainPayload{
		Difficulty: bc.difficulty,
		Blocks:     make([]Block, len(bc.blocks)),
		Pending:    make([]Transaction, len(bc.pending)),
	}
	for i, b := range bc.blocks {
		bcp.Blocks[i] = *b
	}
	for i, t := range bc.pending {
		bcp.Pending[i] = *t
	}
	return bcp
}

func pushToNode(node Node, payload []byte) error {
	_, err := httputil.Put(node.chainUrl(), payload)
	if err.StatusCode == http.StatusConflict {
		return errChainOutdated
	}
	return err
}

func pullFromNode(node Node, bc *Blockchain) error {
	data, err := httputil.Get(node.chainUrl())
	if err != nil {
		return err
	}

	bcp := blockchainPayload{}
	if err := bcp.unmarshal(data); err != nil {
		return err
	}

	muChainAccept.Lock()
	defer muChainAccept.Unlock()

	if len(bc.blocks) > len(bcp.Blocks) {
		bc.setPayload(&bcp)
		return nil
	}
	return errChainOutdated
}

func (bc *Blockchain) SyncWithAdjacentNodes(mode syncMode) error {
	data, err := bc.getPayload().marshal()
	if err != nil {
		return err
	}

	syncPushFirst := func(n Node) {
		if err := pushToNode(n, data); err == errChainOutdated {
			pullFromNode(n, bc)
		}
	}
	syncPullFirst := func(n Node) {
		if err := pullFromNode(n, bc); err == errChainOutdated {
			pushToNode(n, data)
		}
	}

	for _, n := range bc.adjacentNodes {
		if mode == SyncModePushFirst {
			go syncPushFirst(n)
		} else if mode == SyncModePullFirst {
			go syncPullFirst(n)
		} else {
			panic(fmt.Sprintf("invalid blockhain sync mode %d", mode))
		}
	}
	return nil
}

func (bc *Blockchain) RegisterAdjacentNode(node Node) error {
	if err := pullFromNode(node, bc); err != nil {
		if err == errChainOutdated {
			var data []byte
			if data, err = bc.getPayload().marshal(); err == nil {
				err = pushToNode(node, data)
			}
		}
		if err != nil {
			return err
		}
	}
	bc.adjacentNodes = append(bc.adjacentNodes, node)
	return nil
}
