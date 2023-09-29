package indexer

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewOperationHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		payloadProc, err := NewPayloadProcessor(&testscommon.IndexerStub{}, &testscommon.MarshallerMock{})
		require.False(t, payloadProc.IsInterfaceNil())
		require.NotNil(t, payloadProc)
		require.Nil(t, err)

		err = payloadProc.Close()
		require.Nil(t, err)
	})

	t.Run("nil indexer, should return error", func(t *testing.T) {
		payloadProc, err := NewPayloadProcessor(nil, &testscommon.MarshallerMock{})
		require.Nil(t, payloadProc)
		require.Equal(t, errNilIndexer, err)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		payloadProc, err := NewPayloadProcessor(&testscommon.IndexerStub{}, nil)
		require.Nil(t, payloadProc)
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

		payloadProc, _ := NewPayloadProcessor(indexerStub, marshaller)

		err := payloadProc.ProcessPayload(blockBytes, outport.TopicSaveBlock, 0)
		require.True(t, saveBlockCalled)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("invalid bytes"), outport.TopicSaveBlock, 0)
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

		payloadProc, _ := NewPayloadProcessor(indexerStub, marshaller)

		err := payloadProc.ProcessPayload(blockBytes, outport.TopicFinalizedBlock, 0)
		require.True(t, finalizedBlockCalled)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("invalid bytes"), outport.TopicFinalizedBlock, 0)
		require.NotNil(t, err)
	})

	t.Run("no operation handlers", func(t *testing.T) {
		t.Parallel()

		payloadProc, _ := NewPayloadProcessor(&testscommon.IndexerStub{}, marshaller)

		err := payloadProc.ProcessPayload([]byte("payload"), outport.TopicRevertIndexedBlock, 0)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("payload"), outport.TopicSaveRoundsInfo, 0)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("payload"), outport.TopicSaveValidatorsRating, 0)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("payload"), outport.TopicSaveValidatorsPubKeys, 0)
		require.Nil(t, err)

		err = payloadProc.ProcessPayload([]byte("payload"), outport.TopicSaveAccounts, 0)
		require.Nil(t, err)
	})

	t.Run("handler not found", func(t *testing.T) {
		t.Parallel()

		payloadProc, _ := NewPayloadProcessor(&testscommon.IndexerStub{}, marshaller)

		err := payloadProc.ProcessPayload([]byte("payload"), "0xFFFFF", 0)
		require.True(t, strings.Contains(err.Error(), errOperationTypeInvalid.Error()))
		require.True(t, strings.Contains(err.Error(), "0xFFFFF"))
	})
}
