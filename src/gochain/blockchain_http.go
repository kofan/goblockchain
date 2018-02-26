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
	difficulty uint8
	blocks     []Block
	pending    []Transaction
}

func (bcp *blockchainPayload) unmarshal(data []byte) error {
	if err := json.Unmarshal(data, bcp); err != nil {
		return fmt.Errorf("cannot parse blockchain data: %v", err)
	}
	return nil
}

func (bcp *blockchainPayload) marshal(bc *Blockchain) ([]byte, error) {
	bcp.blocks = bc.blocks
	bcp.pending = bc.pending
	bcp.difficulty = bc.difficulty
	return json.Marshal(bcp)
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

	if len(bc.blocks) > len(bcp.blocks) {
		bc.blocks = bcp.blocks
		bc.pending = bcp.pending
		bc.difficulty = bcp.difficulty
		return nil
	}
	return errChainOutdated
}

func (bc *Blockchain) SyncWithAdjacentNodes(mode syncMode) error {
	bcp := blockchainPayload{}
	data, err := bcp.marshal(bc)

	if err != nil {
		return err
	}

	syncPushFirst := func(n Node) {
		if err = pushToNode(n, data); err == errChainOutdated {
			pullFromNode(n, bc)
		}
	}
	syncPullFirst := func(n Node) {
		if err = pullFromNode(n, bc); err == errChainOutdated {
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
			var bcp blockchainPayload
			var data []byte

			if data, err = bcp.marshal(bc); err == nil {
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
