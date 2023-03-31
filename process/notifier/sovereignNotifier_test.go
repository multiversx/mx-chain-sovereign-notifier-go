package notifier

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestSovereignNotifier_Notify(t *testing.T) {
	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	args := ArgsSovereignNotifier{
		Marshaller: &testscommon.MarshallerMock{},
		BlockContainer: &testscommon.BlockContainerStub{
			GetCalled: func(headerType core.HeaderType) (block.EmptyBlockCreator, error) {
				return block.NewEmptyHeaderV2Creator(), nil
			},
		},
		SubscribedAddresses: [][]byte{addr1, addr2},
	}
	sn, _ := NewSovereignNotifier(args)

	txHash1 := []byte("hash1")
	txHash2 := []byte("hash2")
	txHash3 := []byte("hash3")
	txHash4 := []byte("hash4")

	headerV2 := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("root hash"),
	}
	handler1 := &testscommon.ExtendedHeaderHandlerStub{
		SaveExtendedHeaderCalled: func(header *block.ShardHeaderExtended) {
			require.Equal(t, &block.ShardHeaderExtended{
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
			}, header)

		},
	}
	handler2 := &testscommon.ExtendedHeaderHandlerStub{
		SaveExtendedHeaderCalled: func(header *block.ShardHeaderExtended) {
			require.Equal(t, &block.ShardHeaderExtended{
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
			}, header)
		},
	}

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
						RcvAddr: []byte("addr3"),
					},
					ExecutionOrder: 0,
				},
			},
		},
	}

	err = sn.Notify(outportBlock)
	require.Nil(t, err)
}
