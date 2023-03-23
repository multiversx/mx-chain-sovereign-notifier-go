package indexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewOperationHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		oh, err := NewOperationHandler(&testscommon.IndexerStub{}, &testscommon.MarshallerMock{})
		require.NotNil(t, oh)
		require.Nil(t, err)
	})

	t.Run("nil indexer, should return error", func(t *testing.T) {
		oh, err := NewOperationHandler(nil, &testscommon.MarshallerMock{})
		require.Nil(t, oh)
		require.Equal(t, errNilIndexer, err)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		oh, err := NewOperationHandler(&testscommon.IndexerStub{}, nil)
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

		oh, _ := NewOperationHandler(indexerStub, marshaller)
		handlerFunc, found := oh.GetOperationHandler(data.OperationSaveBlock)
		require.True(t, found)

		err := handlerFunc(blockBytes)
		require.Nil(t, err)
		require.True(t, saveBlockCalled)

		err = handlerFunc([]byte("invalid bytes"))
		require.NotNil(t, err)
	})

	t.Run("finalized block", func(t *testing.T) {
		t.Parallel()

		block := &outport.FinalizedBlock{HeaderHash: []byte("hash")}
		blockBytes, _ := marshaller.Marshal(block)
		saveBlockCalled := false

		indexerStub := &testscommon.IndexerStub{
			FinalizedBlockCalled: func(finalizedBlock *outport.FinalizedBlock) error {
				saveBlockCalled = true
				require.Equal(t, block, finalizedBlock)
				return nil
			},
		}

		oh, _ := NewOperationHandler(indexerStub, marshaller)
		handlerFunc, found := oh.GetOperationHandler(data.OperationFinalizedBlock)
		require.True(t, found)

		err := handlerFunc(blockBytes)
		require.Nil(t, err)
		require.True(t, saveBlockCalled)

		err = handlerFunc([]byte("invalid bytes"))
		require.NotNil(t, err)
	})

	t.Run("dummy handlers", func(t *testing.T) {
		t.Parallel()

		oh, _ := NewOperationHandler(&testscommon.IndexerStub{}, marshaller)

		handlerFunc, found := oh.GetOperationHandler(data.OperationRevertIndexedBlock)
		require.True(t, found)
		err := handlerFunc([]byte("bytes"))
		require.Nil(t, err)

		handlerFunc, found = oh.GetOperationHandler(data.OperationSaveRoundsInfo)
		require.True(t, found)
		err = handlerFunc([]byte("bytes"))
		require.Nil(t, err)

		handlerFunc, found = oh.GetOperationHandler(data.OperationSaveValidatorsRating)
		require.True(t, found)
		err = handlerFunc([]byte("bytes"))
		require.Nil(t, err)

		handlerFunc, found = oh.GetOperationHandler(data.OperationSaveValidatorsPubKeys)
		require.True(t, found)
		err = handlerFunc([]byte("bytes"))
		require.Nil(t, err)

		handlerFunc, found = oh.GetOperationHandler(data.OperationSaveAccounts)
		require.True(t, found)
		err = handlerFunc([]byte("bytes"))
		require.Nil(t, err)
	})

	t.Run("handler not found", func(t *testing.T) {
		t.Parallel()

		oh, _ := NewOperationHandler(&testscommon.IndexerStub{}, marshaller)

		handlerFunc, found := oh.GetOperationHandler(data.OperationTypeFromUint64(0xFFFFF))
		require.False(t, found)
		require.Nil(t, handlerFunc)
	})
}
