package indexer

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type handlerFunc func(marshalledData []byte) error

type payloadProcessor struct {
	indexer           process.Indexer
	marshaller        marshal.Marshalizer
	operationHandlers map[string]handlerFunc
}

// NewPayloadProcessor creates a new operation handler
func NewPayloadProcessor(indexer process.Indexer, marshaller marshal.Marshalizer) (DataProcessor, error) {
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

	opHandler.operationHandlers = map[string]handlerFunc{
		outport.TopicSaveBlock:             opHandler.saveBlock,
		outport.TopicRevertIndexedBlock:    noOpHandler,
		outport.TopicSaveRoundsInfo:        noOpHandler,
		outport.TopicSaveValidatorsRating:  noOpHandler,
		outport.TopicSaveValidatorsPubKeys: noOpHandler,
		outport.TopicSaveAccounts:          noOpHandler,
		outport.TopicFinalizedBlock:        opHandler.finalizedBlock,
	}

	return opHandler, nil
}

// ProcessPayload executes the handler func that will index data for requested operation type, if exists
func (pp *payloadProcessor) ProcessPayload(payload []byte, topic string) error {
	handlerFunc, found := pp.operationHandlers[topic]
	if !found {
		return fmt.Errorf("%w, operation type = %s", errOperationTypeInvalid, topic)
	}

	return handlerFunc(payload)
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
