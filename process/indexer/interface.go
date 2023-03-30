package indexer

import "github.com/multiversx/mx-chain-core-go/data/outport"

// LRUOutportBlockCache defines a simple LRU cache able to store *outport.OutportBlock
type LRUOutportBlockCache interface {
	Add(outportBlock *outport.OutportBlock) error
	Get(headerHash []byte) (*outport.OutportBlock, error)
	IsInterfaceNil() bool
}
