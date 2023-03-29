package indexer

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type payloadProcessor struct {
	indexer           process.Indexer
	marshaller        marshal.Marshalizer
	operationHandlers map[data.OperationType]client.HandlerFunc
}

// NewPayloadProcessor creates a new operation handler
func NewPayloadProcessor(indexer process.Indexer, marshaller marshal.Marshalizer) (client.PayloadProcessor, error) {
	if check.IfNil(marshaller) {
		return nil, errNilMarshaller
	}
	if check.IfNil(indexer) {
		return nil, errNilIndexer
	}

	opHandler := &payloadProcessor{
		indexer:    indexer,
		marshaller: marshaller,
	}

	opHandler.operationHandlers = map[data.OperationType]client.HandlerFunc{
		data.OperationSaveBlock:             opHandler.saveBlock,
		data.OperationRevertIndexedBlock:    noOpHandler,
		data.OperationSaveRoundsInfo:        noOpHandler,
		data.OperationSaveValidatorsRating:  noOpHandler,
		data.OperationSaveValidatorsPubKeys: noOpHandler,
		data.OperationSaveAccounts:          noOpHandler,
		data.OperationFinalizedBlock:        opHandler.finalizedBlock,
	}

	return opHandler, nil
}

// ProcessPayload executes the handler func that will index data for requested operation type, if exists
func (pp *payloadProcessor) ProcessPayload(payload *data.PayloadData) error {
	handlerFunc, found := pp.operationHandlers[payload.OperationType]
	if !found {
		return fmt.Errorf("%w, operation type = %s", errOperationTypeInvalid, payload.OperationType.String())
	}

	return handlerFunc(payload.Payload)
}

func (pp *payloadProcessor) saveBlock(marshalledData []byte) error {
	outportBlock := &outport.OutportBlock{}
	err := pp.marshaller.Unmarshal(outportBlock, marshalledData)
	if err != nil {
		return err
	}

	return pp.indexer.SaveBlock(outportBlock)
}

func noOpHandler(_ []byte) error {
	return nil
}

func (pp *payloadProcessor) finalizedBlock(marshalledData []byte) error {
	finalizedBlock := &outport.FinalizedBlock{}
	err := pp.marshaller.Unmarshal(finalizedBlock, marshalledData)
	if err != nil {
		return err
	}

	return pp.indexer.FinalizedBlock(finalizedBlock)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (pp *payloadProcessor) IsInterfaceNil() bool {
	return pp == nil
}

// Close does nothing for now
func (pp *payloadProcessor) Close() error {
	return nil
}
