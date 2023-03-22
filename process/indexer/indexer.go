package indexer

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type indexer struct {
	notifier process.SovereignNotifier
}

// NewIndexer creates an indexer which will internally save received blocks
// and notify sovereign shards for each finalized block
func NewIndexer(notifier process.SovereignNotifier) (process.Indexer, error) {
	if notifier == nil {
		return nil, errNilSovereignNotifier
	}

	return &indexer{
		notifier: notifier,
	}, nil
}

// SaveBlock will save the received block in an internal cache
func (i *indexer) SaveBlock(_ *outport.OutportBlock) error {
	return nil
}

// FinalizedBlock will check the finalized header for incoming txs
// to sovereign shard and push the finalized block through notifier
func (i *indexer) FinalizedBlock(_ *outport.FinalizedBlock) error {
	return nil
}

func (i *indexer) IsInterfaceNil() bool {
	return i == nil
}
