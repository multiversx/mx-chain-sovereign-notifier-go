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

var identifier = []byte("deposit")

func createArgs() ArgsSovereignNotifier {
	return ArgsSovereignNotifier{
		Marshaller: &testscommon.MarshallerMock{},
		SubscribedEvents: []SubscribedEvent{
			{
				Identifier: identifier,
				Addresses: map[string]string{
					"encodedAddr": "decodedAddr",
				},
			},
		},
		Hasher: sha256.NewSha256(),
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
		args.SubscribedEvents = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errNoSubscribedEvent, err)
		require.Nil(t, notif)
	})

	t.Run("no subscribed identifier, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents[0].Identifier = nil
		notif, err := NewSovereignNotifier(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoSubscribedIdentifier.Error()))
		require.True(t, strings.Contains(err.Error(), "index = 0"))
		require.Nil(t, notif)
	})

	t.Run("no subscribed address, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents[0].Addresses = nil
		notif, err := NewSovereignNotifier(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoSubscribedAddresses.Error()))
		require.True(t, strings.Contains(err.Error(), "index = 0"))
		require.Nil(t, notif)

		args.SubscribedEvents[0].Addresses = map[string]string{
			"addr": "",
		}
		notif, err = NewSovereignNotifier(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoSubscribedAddresses.Error()))
		require.True(t, strings.Contains(err.Error(), "index = 0"))
		require.Nil(t, notif)

		args.SubscribedEvents[0].Addresses = map[string]string{
			"": "addr",
		}
		notif, err = NewSovereignNotifier(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errNoSubscribedAddresses.Error()))
		require.True(t, strings.Contains(err.Error(), "index = 0"))
		require.Nil(t, notif)
	})

}

func TestSovereignNotifier_Notify(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	identifier2 := []byte("send")
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
			{
				Address:    addr3,
				Identifier: identifier2,
				Data:       []byte("data7"),
			},
		},
	}

	args := createArgs()
	args.SubscribedEvents = []SubscribedEvent{
		{
			Identifier: identifier,
			Addresses: map[string]string{
				string(addr1): string(addr1),
				string(addr2): string(addr2),
			},
		},
		{
			Identifier: identifier2,
			Addresses: map[string]string{
				string(addr3): string(addr3),
			},
		},
	}

	extendedShardHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, incomingHeader)
	require.Nil(t, err)

	saveHeaderCalled1 := false
	saveHeaderCalled2 := false
	handler1 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, incomingHeader, header)
			saveHeaderCalled1 = true

			return nil
		},
	}
	handler2 := &testscommon.HeaderSubscriberStub{
		AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
			require.Equal(t, extendedShardHeaderHash, headerHash)
			require.Equal(t, incomingHeader, header)
			saveHeaderCalled2 = true

			return nil
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
				{
					TxHash: "txHash3",
					Log: &transaction.Log{
						Events: []*transaction.Event{
							{
								Address:    addr3,
								Identifier: identifier2,
								Data:       []byte("data7"),
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
		require.Equal(t, errNilHeaderSubscriber, err)
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

	t.Run("subscriber cannot add header", func(t *testing.T) {
		args := createArgs()
		sn, _ := NewSovereignNotifier(args)

		blockData := createBlockData(args.Marshaller)

		outportBlock := &outport.OutportBlock{
			BlockData:       blockData,
			TransactionPool: &outport.TransactionPool{},
		}

		errAddHeader := errors.New("cannot add header")
		subscriber := &testscommon.HeaderSubscriberStub{
			AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
				return errAddHeader
			},
		}
		_ = sn.RegisterHandler(subscriber)

		err := sn.Notify(outportBlock)
		require.Equal(t, errAddHeader, err)
	})
}

func TestSovereignNotifier_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
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
	args.SubscribedEvents = []SubscribedEvent{
		{
			Identifier: identifier,
			Addresses: map[string]string{
				string(addr1): string(addr1),
			},
		},
	}

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
					AddHeaderCalled: func(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
						require.Equal(t, extendedShardHeaderHash, headerHash)
						require.Equal(t, incomingHeader, header)

						return nil
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

	sn.headersNotifier.mutSubscribers.RLock()
	defer sn.headersNotifier.mutSubscribers.RUnlock()
	require.Equal(t, n/2, len(sn.headersNotifier.subscribers))
}
