package testscommon

import "github.com/multiversx/mx-chain-core-go/data/outport"

// LRUOutportBlockCacheStub -
type LRUOutportBlockCacheStub struct {
	AddCalled func(outportBlock *outport.OutportBlock) error
	GetCalled func(headerHash []byte) (*outport.OutportBlock, error)
}

// Add -
func (lru *LRUOutportBlockCacheStub) Add(outportBlock *outport.OutportBlock) error {
	if lru.AddCalled != nil {
		return lru.AddCalled(outportBlock)
	}
	return nil
}

// Get -
func (lru *LRUOutportBlockCacheStub) Get(headerHash []byte) (*outport.OutportBlock, error) {
	if lru.GetCalled != nil {
		return lru.GetCalled(headerHash)
	}

	return nil, nil
}

// IsInterfaceNil -
func (lru *LRUOutportBlockCacheStub) IsInterfaceNil() bool {
	return lru == nil
}
