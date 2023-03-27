package indexer

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type lruOutportBlockCache struct {
	cacheSize  uint64
	hashes     []string
	cache      map[string]*outport.OutportBlock
	cacheMutex sync.RWMutex
}

func NewLRUOutportBlockCache() *lruOutportBlockCache {
	return &lruOutportBlockCache{}
}

func (lru *lruOutportBlockCache) Add(outportBlock *outport.OutportBlock) error {
	lru.cacheMutex.Lock()
	lru.cacheMutex.Unlock()

	if outportBlock == nil || outportBlock.BlockData == nil {
		return fmt.Errorf("nil outport block")
	}

	lru.tryEvictCache()

	headerHash := string(outportBlock.BlockData.HeaderHash)
	lru.cache[headerHash] = outportBlock
	lru.hashes = append(lru.hashes, headerHash)

	return nil
}

func (lru *lruOutportBlockCache) tryEvictCache() {
	numEntriesToRemove := len(lru.hashes) - int(lru.cacheSize)
	if numEntriesToRemove <= 0 {
		return
	}

	hashesToRemove := lru.hashes[:numEntriesToRemove]
	for _, hashToRemove := range hashesToRemove {
		delete(lru.cache, hashToRemove)
	}
}

func (lru *lruOutportBlockCache) Get(headerHash []byte) (*outport.OutportBlock, error) {
	headerHashStr := string(headerHash)

	lru.cacheMutex.RLock()
	outportBlock, exists := lru.cache[headerHashStr]
	lru.cacheMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("not found")
	}

	return outportBlock, nil
}
