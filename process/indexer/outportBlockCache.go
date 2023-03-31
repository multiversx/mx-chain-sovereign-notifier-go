package indexer

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type outportBlockCache struct {
	cache      map[string]*outport.OutportBlock
	cacheMutex sync.RWMutex
}

// NewOutportBlockCache creates a new LRU cache able to store *outport.OutportBlock
func NewOutportBlockCache() *outportBlockCache {
	return &outportBlockCache{
		cache:      make(map[string]*outport.OutportBlock),
		cacheMutex: sync.RWMutex{},
	}
}

// Add will add the block to internal cache, if not nil and if the hash doesn't already exist
// If full, it will evict the oldest entry in cache
func (obc *outportBlockCache) Add(outportBlock *outport.OutportBlock) error {
	obc.cacheMutex.Lock()
	defer obc.cacheMutex.Unlock()

	if outportBlock == nil || outportBlock.BlockData == nil {
		return errNilOutportBlock
	}

	hash := outportBlock.BlockData.HeaderHash
	hashStr := string(hash)
	_, exists := obc.cache[hashStr]
	if exists {
		return fmt.Errorf("%w, hash: %s", errOutportBlockAlreadyExists, hex.EncodeToString(hash))
	}

	obc.cache[hashStr] = outportBlock

	return nil
}

// Extract will extract the outport block specified by the header hash, if exists. Otherwise, returns error
func (obc *outportBlockCache) Extract(headerHash []byte) (*outport.OutportBlock, error) {
	hashStr := string(headerHash)

	obc.cacheMutex.RLock()
	outportBlock, exists := obc.cache[hashStr]
	obc.cacheMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w for header hash: %s",
			errOutportBlockNotFound, hex.EncodeToString(headerHash))
	}

	obc.cacheMutex.Lock()
	delete(obc.cache, hashStr)
	obc.cacheMutex.Unlock()

	return outportBlock, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (obc *outportBlockCache) IsInterfaceNil() bool {
	return obc == nil
}
