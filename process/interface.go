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

type Indexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	IsInterfaceNil() bool
}

type HandlerFunc func(d []byte) error

type OperationHandler interface {
	GetOperationsMap() map[data.OperationType]HandlerFunc
}
