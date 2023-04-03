package notifier

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsSovereignNotifier {
	return ArgsSovereignNotifier{
		Marshaller:          &testscommon.MarshallerMock{},
		SubscribedAddresses: [][]byte{[]byte("address")},
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

	t.Run("no subscribed address, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedAddresses = nil
		notif, err := NewSovereignNotifier(args)
		require.Equal(t, errNoSubscribedAddresses, err)
		require.Nil(t, notif)
	})
}

func TestSovereignNotifier_Notify(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	txHash1 := []byte("hash1")
	txHash2 := []byte("hash2")
	txHash3 := []byte("hash3")
	txHash4 := []byte("hash4")

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	extendedShardHeader := &block.ShardHeaderExtended{
		Header: headerV2,
		IncomingMiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash2, txHash3, txHash1},
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            block.TxBlock,
				Reserved:        nil,
			},
		},
	}

	saveHeaderCalled1 := false
	saveHeaderCalled2 := false
	handler1 := &testscommon.ExtendedHeaderHandlerStub{
		SaveExtendedHeaderCalled: func(header *block.ShardHeaderExtended) {
			require.Equal(t, extendedShardHeader, header)
			saveHeaderCalled1 = true
		},
	}
	handler2 := &testscommon.ExtendedHeaderHandlerStub{
		SaveExtendedHeaderCalled: func(header *block.ShardHeaderExtended) {
			require.Equal(t, extendedShardHeader, header)
			saveHeaderCalled2 = true
		},
	}

	args := createArgs()
	args.SubscribedAddresses = [][]byte{addr1, addr2}

	sn, _ := NewSovereignNotifier(args)
	_ = sn.RegisterHandler(handler1)
	_ = sn.RegisterHandler(handler2)

	headerBytes, err := args.Marshaller.Marshal(headerV2)
	require.Nil(t, err)

	outportBlock := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderBytes: headerBytes,
			HeaderType:  string(core.ShardHeaderV2),
		},
		TransactionPool: &outport.TransactionPool{
			Transactions: map[string]*outport.TxInfo{
				hex.EncodeToString(txHash1): {
					Transaction: &transaction.Transaction{
						RcvAddr: addr1,
					},
					ExecutionOrder: 3,
				},
				hex.EncodeToString(txHash2): {
					Transaction: &transaction.Transaction{
						RcvAddr: addr1,
					},
					ExecutionOrder: 1,
				},
				hex.EncodeToString(txHash3): {
					Transaction: &transaction.Transaction{
						RcvAddr: addr2,
					},
					ExecutionOrder: 2,
				},
				hex.EncodeToString(txHash4): {
					Transaction: &transaction.Transaction{
						RcvAddr: addr3,
					},
					ExecutionOrder: 0,
				},
			},
		},
	}

	err = sn.Notify(outportBlock)
	require.Nil(t, err)
	require.True(t, saveHeaderCalled1)
	require.True(t, saveHeaderCalled2)
}

func TestSovereignNotifier_NotifyErrorCases(t *testing.T) {
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
		require.Equal(t, errNilOutportblock, err)

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

	t.Run("notify invalid header typer", func(t *testing.T) {
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
		require.True(t, strings.Contains(err.Error(), errReceivedHeaderType.Error()))
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
}