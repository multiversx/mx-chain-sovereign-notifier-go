package process

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
	Notify(finalizedBlock *outport.OutportBlock) error
	RegisterHandler(handler ExtendedHeaderHandler) error
	IsInterfaceNil() bool
}

// ShardCoordinator should be able to compute an address' shard id
type ShardCoordinator interface {
	ComputeId(address []byte) uint32
	IsInterfaceNil() bool
}

// ExtendedHeaderHandler defines what a subscribed handler to SovereignNotifier should do
type ExtendedHeaderHandler interface {
	AddHeader(headerHash []byte, header data.HeaderHandler)
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
