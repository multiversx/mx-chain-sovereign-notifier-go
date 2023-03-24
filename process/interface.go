package process

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
}

// WSClient defines what a websocket client should do
type WSClient interface {
	Start()
	Close()
}

// Indexer should handle node indexer events
type Indexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	IsInterfaceNil() bool
}
