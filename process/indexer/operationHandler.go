package indexer

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type operationHandler struct {
	indexer    process.Indexer
	marshaller marshal.Marshalizer
}

// NewOperationHandler creates a new operation handler
func NewOperationHandler(indexer process.Indexer, marshaller marshal.Marshalizer) (process.OperationHandler, error) {
	if check.IfNil(marshaller) {
		return nil, errNilMarshaller
	}
	if check.IfNil(indexer) {
		return nil, errNilIndexer
	}

	return &operationHandler{
		indexer:    indexer,
		marshaller: marshaller,
	}, nil
}

// GetOperationsMap returns the map with all the operations that will index data
func (oh *operationHandler) GetOperationsMap() map[data.OperationType]process.HandlerFunc {
	return map[data.OperationType]process.HandlerFunc{
		data.OperationSaveBlock:             oh.saveBlock,
		data.OperationRevertIndexedBlock:    oh.dummyHandler,
		data.OperationSaveRoundsInfo:        oh.dummyHandler,
		data.OperationSaveValidatorsRating:  oh.dummyHandler,
		data.OperationSaveValidatorsPubKeys: oh.dummyHandler,
		data.OperationSaveAccounts:          oh.dummyHandler,
		data.OperationFinalizedBlock:        oh.finalizedBlock,
	}
}

func (oh *operationHandler) saveBlock(marshalledData []byte) error {
	outportBlock := &outport.OutportBlock{}
	err := oh.marshaller.Unmarshal(outportBlock, marshalledData)
	if err != nil {
		return err
	}

	return oh.indexer.SaveBlock(outportBlock)
}

func (oh *operationHandler) dummyHandler(_ []byte) error {
	return nil
}

func (oh *operationHandler) finalizedBlock(marshalledData []byte) error {
	finalizedBlock := &outport.FinalizedBlock{}
	err := oh.marshaller.Unmarshal(finalizedBlock, marshalledData)
	if err != nil {
		return err
	}

	return oh.indexer.FinalizedBlock(finalizedBlock)
}
