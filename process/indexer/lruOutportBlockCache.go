package indexer

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type lruOutportBlockCache struct {
	cacheSize  uint32
	hashes     []string
	cache      map[string]*outport.OutportBlock
	cacheMutex sync.RWMutex
}

// NewLRUOutportBlockCache creates a new LRU cache able to store *outport.OutportBlock
func NewLRUOutportBlockCache(cacheSize uint32) (*lruOutportBlockCache, error) {
	if cacheSize == 0 {
		return nil, errInvalidBlockCacheSize
	}

	return &lruOutportBlockCache{
		cacheSize:  cacheSize,
		hashes:     make([]string, 0, cacheSize),
		cache:      make(map[string]*outport.OutportBlock, cacheSize),
		cacheMutex: sync.RWMutex{},
	}, nil
}

// Add will add the block to internal cache, if not nil and if the hash doesn't already exist
// If full, it will evict the oldest entry in cache
func (lru *lruOutportBlockCache) Add(outportBlock *outport.OutportBlock) error {
	lru.cacheMutex.Lock()
	defer lru.cacheMutex.Unlock()

	if outportBlock == nil || outportBlock.BlockData == nil {
		return errNilOutportBlock
	}

	hash := outportBlock.BlockData.HeaderHash
	hashStr := string(hash)
	_, exists := lru.cache[hashStr]
	if exists {
		return fmt.Errorf("%w, hash: %s", errOutportBlockAlreadyExists, hex.EncodeToString(hash))
	}

	lru.tryEvictCache()

	lru.cache[hashStr] = outportBlock
	lru.hashes = append(lru.hashes, hashStr)

	return nil
}

func (lru *lruOutportBlockCache) tryEvictCache() {
	numEntriesToRemove := len(lru.hashes) - int(lru.cacheSize) + 1
	if numEntriesToRemove <= 0 {
		return
	}

	hashesToRemove := lru.hashes[:numEntriesToRemove]
	lru.hashes = lru.hashes[numEntriesToRemove:]
	for _, hashToRemove := range hashesToRemove {
		delete(lru.cache, hashToRemove)
	}
}

// Get will return a copy of the block specified by the current header hash, if exists. Otherwise, returns error
func (lru *lruOutportBlockCache) Get(headerHash []byte) (*outport.OutportBlock, error) {
	lru.cacheMutex.RLock()
	outportBlock, exists := lru.cache[string(headerHash)]
	lru.cacheMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w for header hash: %s",
			errOutportBlockNotFound, hex.EncodeToString(headerHash))
	}

	outportBlockCopy := *outportBlock
	return &outportBlockCopy, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (lru *lruOutportBlockCache) IsInterfaceNil() bool {
	return lru == nil
}