package process

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
	Notify(finalizedBlock *outport.OutportBlock) error
	RegisterHandler(handler HeaderSubscriber) error
	IsInterfaceNil() bool
}

// ShardCoordinator should be able to compute an address' shard id
type ShardCoordinator interface {
	ComputeId(address []byte) uint32
	IsInterfaceNil() bool
}

// HeaderSubscriber defines a subscriber to incoming headers
type HeaderSubscriber interface {
	AddHeader(headerHash []byte, header data.HeaderHandler)
	IsInterfaceNil() bool
}

// WSClient defines what a websocket client should do
type WSClient interface {
	Start()
	Close() error
}

// Indexer should handle node indexer events
type Indexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	IsInterfaceNil() bool
}
