package notifier

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsSovereignNotifier {
	return ArgsSovereignNotifier{
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

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	identifier := []byte("deposit")

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	incomingHeader := &sovereign.IncomingHeader{
		Header: headerV2,
		IncomingEvents: []*transaction.Event{
			{
				Address:    addr1,
				Identifier: identifier,
				Data:       []byte("data2"),
			},
			{
				Address:    addr2,
				Identifier: identifier,
				Data:       []byte("data5"),
			},
			{
				Address:    addr2,
				Identifier: identifier,
				Data:       []byte("data6"),
			},
		},
	}

	args := createArgs()
	args.SubscribedAddresses = [][]byte{addr1, addr2}

	extendedShardHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, incomingHeader)
	require.Nil(t, err)

	saveHeaderCalled1 := false
	saveHeaderCalled2 := false
	handler1 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, incomingHeader, header)
			saveHeaderCalled1 = true
		},
	}
	handler2 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, incomingHeader, header)
			saveHeaderCalled2 = true
		},
	}

	sn, _ := NewSovereignNotifier(args)
	_ = sn.RegisterHandler(handler1)
	_ = sn.RegisterHandler(handler2)

	headerBytes, err := args.Marshaller.Marshal(headerV2)
	require.Nil(t, err)

	outportBlock := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderHash:  []byte("header hash"),
			HeaderBytes: headerBytes,
			HeaderType:  string(core.ShardHeaderV2),
		},
		TransactionPool: &outport.TransactionPool{
			Logs: []*outport.LogData{
				{
					TxHash: "txHash1",
					Log: &transaction.Log{
						Events: []*transaction.Event{
							{
								Address:    []byte("erd1a"),
								Identifier: []byte("id1"),
								Data:       []byte("data1"),
							},
							{
								Address:    addr1,
								Identifier: identifier,
								Data:       []byte("data2"),
							},
							{
								Address:    addr1,
								Identifier: []byte("id"),
								Data:       []byte("data3"),
							},
						},
					},
				},
				{
					TxHash: "txHash2",
					Log: &transaction.Log{
						Events: []*transaction.Event{
							{
								Address:    []byte("erd1b"),
								Identifier: identifier,
								Data:       []byte("data4"),
							},
							{
								Address:    addr2,
								Identifier: identifier,
								Data:       []byte("data5"),
							},
						},
					},
				},
				{
					TxHash: "txHash2",
					Log: &transaction.Log{
						Events: []*transaction.Event{
							{
								Address:    addr2,
								Identifier: identifier,
								Data:       []byte("data6"),
							},
						},
					},
				},
			},
		},
	}

	err = sn.Notify(outportBlock)
	require.Nil(t, err)
	require.True(t, saveHeaderCalled1)
	require.True(t, saveHeaderCalled2)
}

func TestSovereignNotifier_NotifyRegisterHandlerErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("register invalid extended header handler", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		err := sn.RegisterHandler(nil)
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
	identifier := []byte("deposit")

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	incomingHeader := &sovereign.IncomingHeader{
		Header: headerV2,
		IncomingEvents: []*transaction.Event{
			{
				Address:    addr1,
				Identifier: identifier,
			},
		},
	}

	args := createArgs()
	args.SubscribedAddresses = [][]byte{addr1}

	extendedShardHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, incomingHeader)
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
			Logs: []*outport.LogData{
				{
					TxHash: "txHash",
					Log: &transaction.Log{
						Events: []*transaction.Event{
							{
								Address:    addr1,
								Identifier: identifier,
							},
						},
					},
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
					AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) {
						require.Equal(t, extendedShardHeaderHash, headerHash)
						require.Equal(t, incomingHeader, header)
					},
				}

				errRegister := sn.RegisterHandler(handler)
				require.Nil(t, errRegister)
			}()
		default:
			require.Fail(t, "should not have entered here")
		}
	}

	wg.Wait()

	sn.mutHandler.RLock()
	defer sn.mutHandler.RUnlock()
	require.Equal(t, n/2, len(sn.handlers))
}
