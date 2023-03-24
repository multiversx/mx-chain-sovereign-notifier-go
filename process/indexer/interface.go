package indexer

import "github.com/multiversx/mx-chain-core-go/data/outport"

type LRUOutportBlockCache interface {
	Add(outportBlock *outport.OutportBlock)
	Get(headerHash []byte) (*outport.OutportBlock, error)
}
