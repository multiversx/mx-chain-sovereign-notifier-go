package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
	Notify(finalizedBlock *outport.OutportBlock) error
	RegisterHandler(handler ExtendedHeaderHandler) error
	IsInterfaceNil() bool
}

// ExtendedHeaderHandler defines what a subscribed handler to SovereignNotifier should do
type ExtendedHeaderHandler interface {
	SaveExtendedHeader(header *block.ShardHeaderExtended)
}

// BlockContainerHandler defines what a block container should be able to do
type BlockContainerHandler interface {
	Get(headerType core.HeaderType) (block.EmptyBlockCreator, error)
	IsInterfaceNil() bool
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
