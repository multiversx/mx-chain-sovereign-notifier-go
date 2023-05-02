package notifier

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsSovereignNotifier {
	return ArgsSovereignNotifier{
		ShardCoordinator:    &testscommon.ShardCoordinatorStub{},
		Marshaller:          &testscommon.MarshallerMock{},
		SubscribedAddresses: [][]byte{[]byte("address")},
		Hasher:              sha256.NewSha256(),
	}
}

func createBlockData(marshaller marshal.Marshalizer) *outport.BlockData {
	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}

	headerBytes, _ := marshaller.Marshal(headerV2)
	return &outport.BlockData{
		HeaderBytes: headerBytes,
		HeaderType:  string(core.ShardHeaderV2),
	}
}

func TestNewSovereignNotifier(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		notif, err := NewSovereignNotifier(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(notif))
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgs()
		args.Marshaller = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, notif)
	})

	t.Run("nil shard coordinator, should return error", func(t *testing.T) {
		args := createArgs()
		args.ShardCoordinator = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errNilShardCoordinator, err)
		require.Nil(t, notif)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgs()
		args.Hasher = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errNilHasher, err)
		require.Nil(t, notif)
	})

	t.Run("no subscribed address, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedAddresses = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errNoSubscribedAddresses, err)
		require.Nil(t, notif)
	})

	t.Run("duplicate subscribed address, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedAddresses = [][]byte{[]byte("addr"), []byte("addr")}
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errDuplicateSubscribedAddresses, err)
		require.Nil(t, notif)
	})
}

func TestSovereignNotifier_Notify(t *testing.T) {
	t.Parallel()

	sender1 := []byte("sender1")
	sender2 := []byte("sender2")
	sender3 := []byte("sender3")

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	txHash1 := []byte("hash1")
	txHash2 := []byte("hash2")
	txHash3 := []byte("hash3")
	txHash4 := []byte("hash4")
	txHash5 := []byte("hash5")
	txHash6 := []byte("hash6")

	txs := []*transaction.Transaction{
		{
			RcvAddr: addr1,
			SndAddr: sender1,
		},
		{
			RcvAddr: addr1,
			SndAddr: sender2,
		},
		{
			RcvAddr: addr2,
			SndAddr: sender1,
		},
		{
			RcvAddr: addr3,
			SndAddr: sender2,
		},
		{
			RcvAddr: addr1,
			SndAddr: sender3,
		},
		{
			RcvAddr: addr1,
			SndAddr: sender3,
		},
	}

	incomingOrderedTxs := map[int]*txHandlerInfo{
		0: {
			tx:            txs[1],
			senderShardID: 0,
			hash:          txHash2,
		},
		1: {
			tx:            txs[2],
			senderShardID: 0,
			hash:          txHash3,
		},
		2: {
			tx:            txs[0],
			senderShardID: 0,
			hash:          txHash1,
		},
		3: {
			tx:            txs[4],
			senderShardID: 0,
			hash:          txHash5,
		},
		4: {
			tx:            txs[5],
			senderShardID: 0,
			hash:          txHash6,
		},
	}

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	extendedShardHeader := &block.ShardHeaderExtended{
		Header: headerV2,
		IncomingMiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash2, txHash3, txHash1},
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
				Type:            block.TxBlock,
				Reserved:        []byte(fmt.Sprintf("%d", 0)),
			},
			{
				TxHashes:        [][]byte{txHash5, txHash6},
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
				Type:            block.TxBlock,
				Reserved:        []byte(fmt.Sprintf("%d", 1)),
			},
		},
	}

	args := createArgs()
	args.SubscribedAddresses = [][]byte{addr1, addr2}

	extendedShardHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, extendedShardHeader)
	require.Nil(t, err)

	saveHeaderCalled1 := false
	saveHeaderCalled2 := false
	headerSub1 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header data.HeaderHandler) {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, extendedShardHeader, header)
			saveHeaderCalled1 = true
		},
	}
	headerSub2 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header data.HeaderHandler) {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, extendedShardHeader, header)
			saveHeaderCalled2 = true
		},
	}

	txCount1 := 0
	txCount2 := 0
	txSub1 := &testscommon.TransactionSubscriberStub{
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			require.Equal(t, key, incomingOrderedTxs[txCount1].hash)
			require.Equal(t, data, incomingOrderedTxs[txCount1].tx)
			require.Equal(t, sizeInBytes, incomingOrderedTxs[txCount1].tx.Size())

			expectedCacheID := fmt.Sprintf("%d_%d", core.MainChainShardId, core.SovereignChainShardId)
			require.Equal(t, expectedCacheID, cacheId)

			txCount1++
		},
	}
	txSub2 := &testscommon.TransactionSubscriberStub{
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			require.Equal(t, key, incomingOrderedTxs[txCount2].hash)
			require.Equal(t, data, incomingOrderedTxs[txCount2].tx)
			require.Equal(t, sizeInBytes, incomingOrderedTxs[txCount2].tx.Size())

			expectedCacheID := fmt.Sprintf("%d_%d", core.MainChainShardId, core.SovereignChainShardId)
			require.Equal(t, expectedCacheID, cacheId)

			txCount2++
		},
	}

	args.ShardCoordinator = &testscommon.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			switch string(address) {
			case string(sender1), string(sender2):
				return 0
			case string(sender3):
				return 1
			default:
				require.Fail(t, "should have only 3 senders")
				return 0xFF
			}
		},
	}

	sn, _ := NewSovereignNotifier(args)
	_ = sn.RegisterHeaderSubscriber(headerSub1)
	_ = sn.RegisterHeaderSubscriber(headerSub2)

	_ = sn.RegisterTxSubscriber(txSub1)
	_ = sn.RegisterTxSubscriber(txSub2)

	headerBytes, err := args.Marshaller.Marshal(headerV2)
	require.Nil(t, err)

	outportBlock := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderHash:  []byte("header hash"),
			HeaderBytes: headerBytes,
			HeaderType:  string(core.ShardHeaderV2),
		},
		TransactionPool: &outport.TransactionPool{
			Transactions: map[string]*outport.TxInfo{
				hex.EncodeToString(txHash1): {
					Transaction:    txs[0],
					ExecutionOrder: 3,
				},
				hex.EncodeToString(txHash2): {
					Transaction:    txs[1],
					ExecutionOrder: 1,
				},
				hex.EncodeToString(txHash3): {
					Transaction:    txs[2],
					ExecutionOrder: 2,
				},
				hex.EncodeToString(txHash4): {
					Transaction:    txs[3],
					ExecutionOrder: 0,
				},
				hex.EncodeToString(txHash5): {
					Transaction:    txs[4],
					ExecutionOrder: 0,
				},
				hex.EncodeToString(txHash6): {
					Transaction:    txs[5],
					ExecutionOrder: 3,
				},
			},
		},
	}

	err = sn.Notify(outportBlock)
	require.Nil(t, err)
	require.True(t, saveHeaderCalled1)
	require.True(t, saveHeaderCalled2)
	require.Equal(t, len(incomingOrderedTxs), txCount1)
	require.Equal(t, len(incomingOrderedTxs), txCount2)
}

func TestSovereignNotifier_NotifyRegisterHandlerErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("register invalid extended header handler", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		err := sn.RegisterHeaderSubscriber(nil)
		require.Equal(t, errNilExtendedHeaderHandler, err)
	})

	t.Run("notify nil outport block fields", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		err := sn.Notify(nil)
		require.Equal(t, errNilOutportBlock, err)

		outportBlock := &outport.OutportBlock{
			BlockData:       createBlockData(args.Marshaller),
			TransactionPool: nil,
		}
		err = sn.Notify(outportBlock)
		require.Equal(t, errNilTransactionPool, err)

		outportBlock = &outport.OutportBlock{
			BlockData:       nil,
			TransactionPool: &outport.TransactionPool{},
		}
		err = sn.Notify(outportBlock)
		require.Equal(t, errNilBlockData, err)
	})

	t.Run("notify invalid tx hash", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		invalidHash := "invalid hash"
		outportBlock := &outport.OutportBlock{
			BlockData: createBlockData(args.Marshaller),
			TransactionPool: &outport.TransactionPool{
				Transactions: map[string]*outport.TxInfo{
					invalidHash: {
						Transaction: &transaction.Transaction{RcvAddr: args.SubscribedAddresses[0]},
					},
				},
			},
		}

		err := sn.Notify(outportBlock)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), invalidHash))
	})

	t.Run("notify invalid header type", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		blockData := createBlockData(args.Marshaller)
		blockData.HeaderType = string(core.ShardHeaderV1)
		outportBlock := &outport.OutportBlock{
			BlockData:       blockData,
			TransactionPool: &outport.TransactionPool{},
		}

		err := sn.Notify(outportBlock)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errInvalidHeaderTypeReceived.Error()))
		require.True(t, strings.Contains(err.Error(), blockData.HeaderType))
	})

	t.Run("notify invalid header bytes", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		blockData := createBlockData(args.Marshaller)
		blockData.HeaderBytes = []byte("invalid bytes")

		outportBlock := &outport.OutportBlock{
			BlockData:       blockData,
			TransactionPool: &outport.TransactionPool{},
		}

		err := sn.Notify(outportBlock)
		require.NotNil(t, err)
	})

	t.Run("cannot compute extended header hash", func(t *testing.T) {
		args := createArgs()

		marshalCt := 0
		errMarshal := errors.New("error marshal")
		args.Marshaller = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				marshalCt++
				switch marshalCt {
				case 1:
					return json.Marshal(obj)
				case 2:
					return nil, errMarshal
				}
				return nil, nil
			},
		}

		sn, _ := NewSovereignNotifier(args)

		outportBlock := &outport.OutportBlock{
			BlockData:       createBlockData(args.Marshaller),
			TransactionPool: &outport.TransactionPool{},
		}

		err := sn.Notify(outportBlock)
		require.Equal(t, errMarshal, err)
		require.Equal(t, 2, marshalCt)
	})
}

func TestSovereignNotifier_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	txHash1 := []byte("hash1")

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	extendedShardHeader := &block.ShardHeaderExtended{
		Header: headerV2,
		IncomingMiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash1},
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
				Type:            block.TxBlock,
				Reserved:        []byte(fmt.Sprintf("%d", 0)),
			},
		},
	}

	args := createArgs()
	args.SubscribedAddresses = [][]byte{addr1}

	extendedShardHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, extendedShardHeader)
	require.Nil(t, err)

	sn, _ := NewSovereignNotifier(args)

	headerBytes, err := args.Marshaller.Marshal(headerV2)
	require.Nil(t, err)

	outportBlock := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderHash:  []byte("hash"),
			HeaderBytes: headerBytes,
			HeaderType:  string(core.ShardHeaderV2),
		},
		TransactionPool: &outport.TransactionPool{
			Transactions: map[string]*outport.TxInfo{
				hex.EncodeToString(txHash1): {
					Transaction: &transaction.Transaction{
						RcvAddr: addr1,
					},
					ExecutionOrder: 0,
				},
			},
		},
	}

	n := 100
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		switch i % 2 {
		case 0:
			go func() {
				errNotify := sn.Notify(outportBlock)
				require.Nil(t, errNotify)
			}()
		case 1:
			wg.Add(1)

			go func() {
				defer wg.Done()

				handler := &testscommon.HeaderSubscriberStub{
					AddHeaderCalled: func(headerHash []byte, header data.HeaderHandler) {
						require.Equal(t, extendedShardHeaderHash, headerHash)
						require.Equal(t, extendedShardHeader, header)
					},
				}

				errRegister := sn.RegisterHeaderSubscriber(handler)
				require.Nil(t, errRegister)
			}()
		default:
			require.Fail(t, "should not have entered here")
		}
	}

	wg.Wait()

	sn.headerNotifier.mutSubscribers.RLock()
	defer sn.headerNotifier.mutSubscribers.RUnlock()
	require.Equal(t, n/2, len(sn.headerNotifier.subscribers))
}
