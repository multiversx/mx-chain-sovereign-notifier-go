package indexer

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewIndexer(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		indx, err := NewIndexer(&testscommon.SovereignNotifierStub{}, &testscommon.OutportBlockCacheStub{})
		require.Nil(t, err)
		require.False(t, check.IfNil(indx))
	})

	t.Run("nil sovereign notifier, should error", func(t *testing.T) {
		indx, err := NewIndexer(nil, &testscommon.OutportBlockCacheStub{})
		require.Equal(t, errNilSovereignNotifier, err)
		require.Nil(t, indx)
	})

	t.Run("nil cache, should error", func(t *testing.T) {
		indx, err := NewIndexer(&testscommon.SovereignNotifierStub{}, nil)
		require.Equal(t, errNilOutportBlockCache, err)
		require.Nil(t, indx)
	})
}

func TestIndexer_SaveBlock(t *testing.T) {
	t.Parallel()

	wasAddCalled := false
	cache := &testscommon.OutportBlockCacheStub{
		AddCalled: func(outportBlock *outport.OutportBlock) error {
			wasAddCalled = true
			return nil
		},
	}
	indx, _ := NewIndexer(&testscommon.SovereignNotifierStub{}, cache)

	err := indx.SaveBlock(&outport.OutportBlock{})
	require.Nil(t, err)
	require.True(t, wasAddCalled)
}

func TestIndexer_FinalizedBlock(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasExtractCalled := false
		hash := []byte("hash")
		outportBlock := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: hash}}
		cache := &testscommon.OutportBlockCacheStub{
			ExtractCalled: func(headerHash []byte) (*outport.OutportBlock, error) {
				wasExtractCalled = true
				require.Equal(t, hash, headerHash)

				return outportBlock, nil
			},
		}

		wasNotifyCalled := false
		notifier := &testscommon.SovereignNotifierStub{
			NotifyCalled: func(finalizedBlock *outport.OutportBlock) error {
				wasNotifyCalled = true
				require.Equal(t, outportBlock, finalizedBlock)

				return nil
			},
		}
		indx, _ := NewIndexer(notifier, cache)

		err := indx.FinalizedBlock(&outport.FinalizedBlock{HeaderHash: hash})
		require.Nil(t, err)
		require.True(t, wasExtractCalled)
		require.True(t, wasNotifyCalled)
	})

	t.Run("error getting block from cache", func(t *testing.T) {
		t.Parallel()

		wasExtractCalled := false
		hash := []byte("hash")
		errGetBlock := errors.New("error getting block")
		cache := &testscommon.OutportBlockCacheStub{
			ExtractCalled: func(headerHash []byte) (*outport.OutportBlock, error) {
				wasExtractCalled = true
				require.Equal(t, hash, headerHash)

				return nil, errGetBlock
			},
		}

		wasNotifyCalled := false
		notifier := &testscommon.SovereignNotifierStub{
			NotifyCalled: func(finalizedBlock *outport.OutportBlock) error {
				wasNotifyCalled = true
				return nil
			},
		}
		indx, _ := NewIndexer(notifier, cache)

		err := indx.FinalizedBlock(&outport.FinalizedBlock{HeaderHash: hash})
		require.Equal(t, errGetBlock, err)
		require.True(t, wasExtractCalled)
		require.False(t, wasNotifyCalled)
	})

}
