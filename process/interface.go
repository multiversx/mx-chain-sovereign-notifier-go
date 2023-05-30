package process

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
	Notify(finalizedBlock *outport.OutportBlock) error
	RegisterHandler(handler IncomingHeaderSubscriber) error
	IsInterfaceNil() bool
}

// ShardCoordinator should be able to compute an address' shard id
type ShardCoordinator interface {
	ComputeId(address []byte) uint32
	IsInterfaceNil() bool
}

// IncomingHeaderSubscriber defines a subscriber to incoming headers
type IncomingHeaderSubscriber interface {
	AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error
	IsInterfaceNil() bool
}

// WSClient defines what a websocket client should do
type WSClient interface {
	Close() error
}

// Indexer should handle node indexer events
type Indexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	IsInterfaceNil() bool
}
