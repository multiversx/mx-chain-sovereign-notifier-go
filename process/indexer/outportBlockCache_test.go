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

func requireErrIsBlockNotFound(t *testing.T, err error, hash []byte) {
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), errOutportBlockNotFound.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(hash)))
}

func TestOutportBlockCache_Add_Get(t *testing.T) {
	t.Parallel()

	cache := NewOutportBlockCache()
	require.False(t, check.IfNil(cache))

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

	rcvBl1, err := cache.Extract(h1)
	require.Nil(t, err)
	require.True(t, bl1 == rcvBl1)

	rcvBl2, err := cache.Extract(h2)
	require.Nil(t, err)
	require.True(t, bl2 == rcvBl2)

	err = cache.Add(bl3)
	require.Nil(t, err)

	err = cache.Add(bl4)
	require.Nil(t, err)

	rcvBl1, err = cache.Extract(h1)
	requireErrIsBlockNotFound(t, err, h1)

	rcvBl2, err = cache.Extract(h2)
	requireErrIsBlockNotFound(t, err, h2)

	require.Equal(t, map[string]*outport.OutportBlock{
		string(h3): bl3,
		string(h4): bl4,
	}, cache.cache)
}

func TestOutportBlockCache_AddErrorCases(t *testing.T) {
	t.Parallel()

	cache := NewOutportBlockCache()

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

	require.Equal(t, map[string]*outport.OutportBlock{
		string(h1): bl1,
	}, cache.cache)
}

func TestOutportBlockCache_AddConcurrentOperations(t *testing.T) {
	t.Parallel()

	cache := NewOutportBlockCache()

	n := 10000
	extraElems := 10

	wg1 := sync.WaitGroup{}
	wg1.Add(1)

	wg2 := sync.WaitGroup{}
	wg2.Add(1)

	go func() {
		for i := 0; i < n; i++ {
			hash := []byte(fmt.Sprintf("%d", i))
			err := cache.Add(&outport.OutportBlock{BlockData: &outport.BlockData{HeaderHash: hash}})
			require.Nil(t, err)
		}
		wg1.Done()
	}()

	go func() {
		for i := 0; i < n-extraElems; i++ {
			hash := []byte(fmt.Sprintf("%d", i))
			_, _ = cache.Extract(hash)
		}
		wg2.Done()
	}()

	wg1.Wait()
	wg2.Wait()

	for i := n - extraElems; i < n; i++ {
		hash := []byte(fmt.Sprintf("%d", i))
		_, err := cache.Extract(hash)
		require.Nil(t, err)
	}
}
