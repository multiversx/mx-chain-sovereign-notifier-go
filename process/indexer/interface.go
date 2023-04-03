package indexer

import "github.com/multiversx/mx-chain-core-go/data/outport"

// OutportBlockCache defines a simple cache able to store *outport.OutportBlock
type OutportBlockCache interface {
	Add(outportBlock *outport.OutportBlock) error
	Extract(headerHash []byte) (*outport.OutportBlock, error)
	IsInterfaceNil() bool
}
