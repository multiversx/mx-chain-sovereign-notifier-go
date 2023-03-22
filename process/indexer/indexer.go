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

	return &indexer{
		notifier: notifier,
	}, nil
}

func (i *indexer) SaveBlock(_ *outport.OutportBlock) error {
	return nil
}

func (i *indexer) FinalizedBlock(_ *outport.FinalizedBlock) error {
	return nil
}

func (i *indexer) IsInterfaceNil() bool {
	return i == nil
}
