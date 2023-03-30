package indexer

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestNewLRUOutportBlockCache(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		cache, err := NewLRUOutportBlockCache(1000)
		require.Nil(t, err)
		require.False(t, check.IfNil(cache))
	})

	t.Run("should work", func(t *testing.T) {
		cache, err := NewLRUOutportBlockCache(0)
		require.Equal(t, errInvalidBlockCacheSize, err)
		require.Nil(t, cache)
	})
}

func requireErrIsBlockNotFound(t *testing.T, err error, hash []byte) {
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), errOutportBlockNotFound.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(hash)))
}

func TestLruOutportBlockCache_Add_Get(t *testing.T) {
	t.Parallel()

	cache, _ := NewLRUOutportBlockCache(2)

	h1 := []byte("h1")
	h2 := []byte("h2")
	h3 := []byte("h3")
	h4 := []byte("h4")

	bl1 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: h1}}
	bl2 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: h2}}
	bl3 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: h3}}
	bl4 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: h4}}

	err := cache.Add(bl1)
	require.Nil(t, err)

	err = cache.Add(bl2)
	require.Nil(t, err)

	rcvBl1, err := cache.Get(h1)
	require.Nil(t, err)
	require.Equal(t, bl1, rcvBl1)   // same structf
	require.False(t, bl1 == rcvBl1) // not the same pointer

	rcvBl2, err := cache.Get(h2)
	require.Nil(t, err)
	require.Equal(t, bl1, rcvBl1)   // same struct
	require.False(t, bl2 == rcvBl2) // not the same pointer

	rcvBl3, err := cache.Get(h3)
	requireErrIsBlockNotFound(t, err, h3)
	require.Nil(t, rcvBl3)

	rcvBl4, err := cache.Get(h4)
	requireErrIsBlockNotFound(t, err, h4)
	require.Nil(t, rcvBl4)

	// Cache is full, adding bl3 will remove bl1
	err = cache.Add(bl3)
	require.Nil(t, err)

	rcvBl3, err = cache.Get(h3)
	require.Nil(t, err)
	require.Equal(t, bl3, rcvBl3)   // same struct
	require.False(t, bl3 == rcvBl3) // not the same pointer

	rcvBl2, err = cache.Get(h2)
	require.Nil(t, err)
	require.Equal(t, bl2, rcvBl2)   // same struct
	require.False(t, bl2 == rcvBl2) // not the same pointer

	rcvBl1, err = cache.Get(h1)
	requireErrIsBlockNotFound(t, err, h1)
	require.Nil(t, rcvBl1)

	// Cache is full, adding bl4 will remove bl2
	err = cache.Add(bl4)
	require.Nil(t, err)

	rcvBl4, err = cache.Get(h4)
	require.Nil(t, err)
	require.Equal(t, bl4, rcvBl4)   // same struct
	require.False(t, bl4 == rcvBl4) // not the same pointer

	rcvBl3, err = cache.Get(h3)
	require.Nil(t, err)
	require.Equal(t, bl3, rcvBl3)   // same struct
	require.False(t, bl3 == rcvBl3) // not the same pointer

	rcvBl2, err = cache.Get(h2)
	requireErrIsBlockNotFound(t, err, h2)
	require.Nil(t, rcvBl2)

	rcvBl1, err = cache.Get(h1)
	requireErrIsBlockNotFound(t, err, h1)
	require.Nil(t, rcvBl1)

	require.Equal(t, []string{string(h3), string(h4)}, cache.hashes)
	require.Equal(t, map[string]*outport.OutportBlock{
		string(h3): bl3,
		string(h4): bl4,
	}, cache.cache)
}

func TestLruOutportBlockCache_AddErrorCases(t *testing.T) {
	t.Parallel()

	cache, _ := NewLRUOutportBlockCache(3)

	err := cache.Add(nil)
	require.Equal(t, errNilOutportBlock, err)

	err = cache.Add(&outport.OutportBlock{BlockData: nil})
	require.Equal(t, errNilOutportBlock, err)

	h1 := []byte("h1")
	bl1 := &outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: h1}}

	err = cache.Add(bl1)
	require.Nil(t, err)

	err = cache.Add(bl1)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errOutportBlockAlreadyExists.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(h1)))

	require.Equal(t, []string{string(h1)}, cache.hashes)
	require.Equal(t, map[string]*outport.OutportBlock{
		string(h1): bl1,
	}, cache.cache)
}

func TestLruOutportBlockCache_AddConcurrentOperations(t *testing.T) {
	t.Parallel()

	cacheSize := uint32(10)
	cache, _ := NewLRUOutportBlockCache(cacheSize)

	n := 10000
	wg1 := sync.WaitGroup{}
	wg1.Add(1)

	wg2 := sync.WaitGroup{}
	wg2.Add(1)

	go func() {
		for i := 0; i < n; i++ {
			hash := []byte(fmt.Sprintf("%d", i))
			_ = cache.Add(&outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: hash}})
		}
		wg1.Done()
	}()

	go func() {
		for i := 0; i < n; i++ {
			hash := []byte(fmt.Sprintf("%d", i))
			_, _ = cache.Get(hash)
		}
		wg2.Done()
	}()

	wg1.Wait()
	wg2.Wait()

	require.Len(t, cache.hashes, int(cacheSize))
	require.Len(t, cache.cache, int(cacheSize))

	for i := n - int(cacheSize); i < n; i++ {
		hash := []byte(fmt.Sprintf("%d", i))
		_, err := cache.Get(hash)
		require.Nil(t, err)
	}
}
