package gochain

import (
	"path"
)

// Node represents a single node in the blockhain decentralized net
type Node struct {
	name    string
	address string
}

func (n *Node) chainUrl() string {
	return path.Join(n.address, "/chain")
}
