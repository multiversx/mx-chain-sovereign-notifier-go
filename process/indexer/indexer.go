package indexer

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type indexer struct {
	notifier process.SovereignNotifier
}

func NewIndexer(notifier process.SovereignNotifier) (process.Indexer, error) {
	if notifier == nil {
		return nil, errNilSovereignNotifier
	}

	return &indexer{}, nil
}

func (i *indexer) SaveBlock(outportBlock *outport.OutportBlock) error {
	return nil
}

func (i *indexer) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	return nil
}

func (i *indexer) IsInterfaceNil() bool {
	return i == nil
}
