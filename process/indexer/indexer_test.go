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
		indx, err := NewIndexer(&testscommon.SovereignNotifierStub{}, &testscommon.LRUOutportBlockCacheStub{})
		require.Nil(t, err)
		require.False(t, check.IfNil(indx))
	})

	t.Run("should work", func(t *testing.T) {
		indx, err := NewIndexer(nil, &testscommon.LRUOutportBlockCacheStub{})
		require.Equal(t, errNilSovereignNotifier, err)
		require.Nil(t, indx)
	})

	t.Run("should work", func(t *testing.T) {
		indx, err := NewIndexer(&testscommon.SovereignNotifierStub{}, nil)
		require.Equal(t, errNilOutportBlockCache, err)
		require.Nil(t, indx)
	})
}

func TestIndexer_SaveBlock(t *testing.T) {
	wasAddCalled := false
	cache := &testscommon.LRUOutportBlockCacheStub{
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

		wasGetCalled := false
		hash := []byte("hash")
		outportBlock := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: hash}}
		cache := &testscommon.LRUOutportBlockCacheStub{
			GetCalled: func(headerHash []byte) (*outport.OutportBlock, error) {
				wasGetCalled = true
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
		require.True(t, wasGetCalled)
		require.True(t, wasNotifyCalled)
	})

	t.Run("error getting block from cache", func(t *testing.T) {
		t.Parallel()

		wasGetCalled := false
		hash := []byte("hash")
		errGetBlock := errors.New("error getting block")
		cache := &testscommon.LRUOutportBlockCacheStub{
			GetCalled: func(headerHash []byte) (*outport.OutportBlock, error) {
				wasGetCalled = true
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
		require.True(t, wasGetCalled)
		require.False(t, wasNotifyCalled)
	})

}
