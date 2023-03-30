package indexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestLruOutportBlockCache_Add_Get(t *testing.T) {
	cache, _ := NewLRUOutportBlockCache(2)

	bl1 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: []byte("h1")}}
	bl2 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: []byte("h2")}}
	bl3 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: []byte("h3")}}
	bl4 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: []byte("h4")}}

	err := cache.Add(bl1)
	require.Nil(t, err)

	err = cache.Add(bl2)
	require.Nil(t, err)

	rcvBl1, err := cache.Get(bl1.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl1, rcvBl1)   // same struct
	require.False(t, bl1 == rcvBl1) // not the same pointer

	rcvBl2, err := cache.Get(bl2.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl1, rcvBl1)   // same struct
	require.False(t, bl2 == rcvBl2) // not the same pointer

	// Cache is now full, adding a new elements will remove the oldest
	err = cache.Add(bl3)
	require.Nil(t, err)

	rcvBl3, err := cache.Get(bl3.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl3, rcvBl3)   // same struct
	require.False(t, bl3 == rcvBl3) // not the same pointer

	rcvBl2, err = cache.Get(bl2.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl2, rcvBl2)   // same struct
	require.False(t, bl2 == rcvBl2) // not the same pointer

	rcvBl1, err = cache.Get(bl1.BlockData.HeaderHash)
	require.NotNil(t, err)
	require.Nil(t, rcvBl1)

	err = cache.Add(bl4)
	require.Nil(t, err)

	rcvBl4, err := cache.Get(bl4.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl4, rcvBl4)   // same struct
	require.False(t, bl4 == rcvBl4) // not the same pointer

	rcvBl3, err = cache.Get(bl3.BlockData.HeaderHash)
	require.Nil(t, err)
	require.Equal(t, bl3, rcvBl3)   // same struct
	require.False(t, bl3 == rcvBl3) // not the same pointer

	rcvBl2, err = cache.Get(bl2.BlockData.HeaderHash)
	require.NotNil(t, err)
	require.Nil(t, rcvBl2)

	rcvBl1, err = cache.Get(bl1.BlockData.HeaderHash)
	require.NotNil(t, err)
	require.Nil(t, rcvBl1)

	require.Equal(t, []string{string(bl3.BlockData.HeaderHash), string(bl4.BlockData.HeaderHash)}, cache.hashes)
	require.Equal(t, map[string]*outport.OutportBlock{
		string(bl3.BlockData.HeaderHash): bl3,
		string(bl4.BlockData.HeaderHash): bl4,
	}, cache.cache)
}
