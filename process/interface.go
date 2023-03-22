package process

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
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

// HandlerFunc defines the func responsible for handling received payload data from node
type HandlerFunc func(d []byte) error

// OperationHandler defines a HandlerFunc for each indexer operation type from node
type OperationHandler interface {
	GetOperationsMap() map[data.OperationType]HandlerFunc
}
