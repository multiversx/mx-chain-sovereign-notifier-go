package indexer

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type operationHandler struct {
	indexer           process.Indexer
	marshaller        marshal.Marshalizer
	operationHandlers map[data.OperationType]process.HandlerFunc
}

// NewOperationHandler creates a new operation handler
func NewOperationHandler(indexer process.Indexer, marshaller marshal.Marshalizer) (process.OperationHandler, error) {
	if check.IfNil(marshaller) {
		return nil, errNilMarshaller
	}
	if check.IfNil(indexer) {
		return nil, errNilIndexer
	}

	opHandler := &operationHandler{
		indexer:    indexer,
		marshaller: marshaller,
	}

	opHandler.operationHandlers = map[data.OperationType]process.HandlerFunc{
		data.OperationSaveBlock:             opHandler.saveBlock,
		data.OperationRevertIndexedBlock:    dummyHandler,
		data.OperationSaveRoundsInfo:        dummyHandler,
		data.OperationSaveValidatorsRating:  dummyHandler,
		data.OperationSaveValidatorsPubKeys: dummyHandler,
		data.OperationSaveAccounts:          dummyHandler,
		data.OperationFinalizedBlock:        opHandler.finalizedBlock,
	}

	return opHandler, nil
}

// GetOperationHandler returns the handler func that will index data for requested operation type
func (oh *operationHandler) GetOperationHandler(operation data.OperationType) (process.HandlerFunc, bool) {
	handlerFunc, found := oh.operationHandlers[operation]
	return handlerFunc, found
}

func (oh *operationHandler) saveBlock(marshalledData []byte) error {
	outportBlock := &outport.OutportBlock{}
	err := oh.marshaller.Unmarshal(outportBlock, marshalledData)
	if err != nil {
		return err
	}

	return oh.indexer.SaveBlock(outportBlock)
}

func dummyHandler(_ []byte) error {
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
