package indexer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewOperationHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		oh, err := NewPayloadProcessor(&testscommon.IndexerStub{}, &testscommon.MarshallerMock{})
		require.False(t, oh.IsInterfaceNil())
		require.NotNil(t, oh)
		require.Nil(t, err)
	})

	t.Run("nil indexer, should return error", func(t *testing.T) {
		oh, err := NewPayloadProcessor(nil, &testscommon.MarshallerMock{})
		require.Nil(t, oh)
		require.Equal(t, errNilIndexer, err)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		oh, err := NewPayloadProcessor(&testscommon.IndexerStub{}, nil)
		require.Nil(t, oh)
		require.Equal(t, errNilMarshaller, err)
	})
}

func TestOperationHandler_GetOperationHandler(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshallerMock{}

	t.Run("save block", func(t *testing.T) {
		t.Parallel()

		block := &outport.OutportBlock{NumberOfShards: 3}
		blockBytes, _ := marshaller.Marshal(block)
		saveBlockCalled := false

		indexerStub := &testscommon.IndexerStub{
			SaveBlockCalled: func(outportBlock *outport.OutportBlock) error {
				saveBlockCalled = true
				require.Equal(t, block, outportBlock)
				return nil
			},
		}

		oh, _ := NewPayloadProcessor(indexerStub, marshaller)

		payload := &data.PayloadData{
			OperationType: data.OperationSaveBlock,
			Payload:       blockBytes,
		}
		err := oh.ProcessPayload(payload)
		require.True(t, saveBlockCalled)
		require.Nil(t, err)

		payload.Payload = []byte("invalid bytes")
		err = oh.ProcessPayload(payload)
		require.NotNil(t, err)
	})

	t.Run("finalized block", func(t *testing.T) {
		t.Parallel()

		block := &outport.FinalizedBlock{HeaderHash: []byte("hash")}
		blockBytes, _ := marshaller.Marshal(block)
		finalizedBlockCalled := false

		indexerStub := &testscommon.IndexerStub{
			FinalizedBlockCalled: func(finalizedBlock *outport.FinalizedBlock) error {
				finalizedBlockCalled = true
				require.Equal(t, block, finalizedBlock)
				return nil
			},
		}

		oh, _ := NewPayloadProcessor(indexerStub, marshaller)

		payload := &data.PayloadData{
			OperationType: data.OperationFinalizedBlock,
			Payload:       blockBytes,
		}
		err := oh.ProcessPayload(payload)
		require.True(t, finalizedBlockCalled)
		require.Nil(t, err)

		payload.Payload = []byte("invalid bytes")
		err = oh.ProcessPayload(payload)
		require.NotNil(t, err)
	})

	t.Run("no operation handlers", func(t *testing.T) {
		t.Parallel()

		oh, _ := NewPayloadProcessor(&testscommon.IndexerStub{}, marshaller)

		err := oh.ProcessPayload(&data.PayloadData{OperationType: data.OperationRevertIndexedBlock})
		require.Nil(t, err)

		err = oh.ProcessPayload(&data.PayloadData{OperationType: data.OperationSaveRoundsInfo})
		require.Nil(t, err)

		err = oh.ProcessPayload(&data.PayloadData{OperationType: data.OperationSaveValidatorsRating})
		require.Nil(t, err)

		err = oh.ProcessPayload(&data.PayloadData{OperationType: data.OperationSaveValidatorsPubKeys})
		require.Nil(t, err)

		err = oh.ProcessPayload(&data.PayloadData{OperationType: data.OperationSaveAccounts})
		require.Nil(t, err)
	})

	t.Run("handler not found", func(t *testing.T) {
		t.Parallel()

		oh, _ := NewPayloadProcessor(&testscommon.IndexerStub{}, marshaller)

		payload := &data.PayloadData{OperationType: data.OperationTypeFromUint64(0xFFFFF)}
		err := oh.ProcessPayload(payload)
		require.True(t, strings.Contains(err.Error(), errOperationTypeInvalid.Error()))
		require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%s", payload.OperationType.String())))
	})
}
