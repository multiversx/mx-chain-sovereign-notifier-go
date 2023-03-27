package indexer

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type indexer struct {
	notifier process.SovereignNotifier
	cache    LRUOutportBlockCache
}

// NewIndexer creates an indexer which will internally save received blocks
// and notify sovereign shards for each finalized block
func NewIndexer(notifier process.SovereignNotifier, cache LRUOutportBlockCache) (process.Indexer, error) {
	if notifier == nil {
		return nil, errNilSovereignNotifier
	}
	if cache == nil {
		return nil, errNilOutportBlockCache
	}

	return &indexer{
		cache:    cache,
		notifier: notifier,
	}, nil
}

// SaveBlock will save the received block in an internal cache
func (i *indexer) SaveBlock(outportBlock *outport.OutportBlock) error {
	return i.cache.Add(outportBlock)
}

// FinalizedBlock will check the finalized header for incoming txs
// to sovereign shard and push the finalized block through notifier
func (i *indexer) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	outportBlock, err := i.cache.Get(finalizedBlock.HeaderHash)
	if err != nil {
		return err
	}

	return i.notifier.Notify(outportBlock)
}

func (i *indexer) IsInterfaceNil() bool {
	return i == nil
}
