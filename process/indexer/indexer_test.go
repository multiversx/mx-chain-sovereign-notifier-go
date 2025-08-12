package indexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
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

	hash := []byte("hash")
	outportBlock := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: hash}}
	wasNotifyCalled := false
	notifier := &testscommon.SovereignNotifierStub{
		NotifyCalled: func(finalizedBlock *outport.OutportBlock) error {
			wasNotifyCalled = true
			require.Equal(t, outportBlock, finalizedBlock)

			return nil
		},
	}
	indx, _ := NewIndexer(notifier, &testscommon.OutportBlockCacheStub{})

	err := indx.SaveBlock(outportBlock)
	require.Nil(t, err)
	require.True(t, wasNotifyCalled)
}
