package notifier

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type ExtendedHeaderHandler func(header *block.ShardHeaderExtended)

type sovereignNotifier struct {
	mutHandler          sync.RWMutex
	handlers            []process.ExtendedHeaderHandler
	subscribedAddresses [][]byte
	headerV2Creator     block.EmptyBlockCreator
	marshaller          marshal.Marshalizer
}

type ArgsSovereignNotifier struct {
	Marshaller          marshal.Marshalizer
	SubscribedAddresses [][]byte
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(args ArgsSovereignNotifier) (*sovereignNotifier, error) {
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	err := checkAddresses(args.SubscribedAddresses)
	if err != nil {
		return nil, err
	}

	log.Debug("received config", "subscribed addresses", args.SubscribedAddresses)

	return &sovereignNotifier{
		subscribedAddresses: args.SubscribedAddresses,
		handlers:            make([]process.ExtendedHeaderHandler, 0),
		mutHandler:          sync.RWMutex{},
		headerV2Creator:     block.NewEmptyHeaderV2Creator(),
		marshaller:          args.Marshaller,
	}, nil
}

func checkAddresses(addresses [][]byte) error {
	numAddresses := len(addresses)
	if numAddresses == 0 {
		return errNoSubscribedAddresses
	}

	addressesMap := make(map[string]struct{})
	for _, addr := range addresses {
		addressesMap[string(addr)] = struct{}{}
	}

	if len(addressesMap) != numAddresses {
		return errDuplicateSubscribedAddresses
	}

	return nil
}

// Notify will notify the sovereign nodes about the finalized block and incoming mb txs
// For each subscribed address, it searches if the receiver is found in tx in pool.
// If found, IncomingMiniBlocks will contain the ordered tx hashes by execution.
func (notifier *sovereignNotifier) Notify(outportBlock *outport.OutportBlock) error {
	err := checkNilOutportBlockFields(outportBlock)
	if err != nil {
		return err
	}

	mbs, err := notifier.getAllIncomingMbs(outportBlock.TransactionPool)
	if err != nil {
		return err
	}

	headerV2, err := notifier.getHeaderV2(core.HeaderType(outportBlock.BlockData.HeaderType), outportBlock.BlockData.HeaderBytes)
	if err != nil {
		return err
	}

	extendedHeader := &block.ShardHeaderExtended{
		Header:             headerV2,
		IncomingMiniBlocks: mbs,
	}

	notifier.notifyHandlers(extendedHeader)
	return nil
}

func checkNilOutportBlockFields(outportBlock *outport.OutportBlock) error {
	if outportBlock == nil {
		return errNilOutportblock
	}
	if outportBlock.TransactionPool == nil {
		return errNilTransactionPool
	}
	if outportBlock.BlockData == nil {
		return errNilBlockData
	}

	return nil
}

func (notifier *sovereignNotifier) getAllIncomingMbs(txPool *outport.TransactionPool) ([]*block.MiniBlock, error) {
	mbs := make([]*block.MiniBlock, 0)

	txsMb, err := notifier.getIncomingMbFromTxs(txPool.Transactions)
	if err != nil {
		return nil, err
	}

	// TODO: when specs are defined, we should also handle scrs mbs
	mbs = append(mbs, txsMb)
	return mbs, nil
}

func (notifier *sovereignNotifier) getIncomingMbFromTxs(txs map[string]*outport.TxInfo) (*block.MiniBlock, error) {
	txHashes := make([][]byte, 0)
	execOrderTxHashMap := make(map[string]uint32)

	for txHash, tx := range txs {
		if !contains(notifier.subscribedAddresses, tx.GetTransaction().GetRcvAddr()) {
			continue
		}

		hashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, fmt.Errorf("%w, hash: %s", err, txHash)
		}

		log.Info("found incoming tx", "tx hash", txHash)

		txHashes = append(txHashes, hashBytes)
		execOrderTxHashMap[string(hashBytes)] = tx.GetExecutionOrder()
	}

	sort.SliceStable(txHashes, func(i, j int) bool {
		return execOrderTxHashMap[string(txHashes[i])] < execOrderTxHashMap[string(txHashes[j])]
	})

	return &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0, // todo: decide what we should fill here
		SenderShardID:   0, // todo: decide what we should fill here
		Type:            block.TxBlock,
		Reserved:        nil,
	}, nil
}

func contains(addresses [][]byte, address []byte) bool {
	for _, addr := range addresses {
		if bytes.Equal(address, addr) {
			return true
		}
	}

	return false
}

func (notifier *sovereignNotifier) getHeaderV2(headerType core.HeaderType, headerBytes []byte) (*block.HeaderV2, error) {
	if headerType != core.ShardHeaderV2 {
		return nil, fmt.Errorf("%w : %s, expected: %s",
			errReceivedHeaderType, headerType, core.ShardHeaderV2)
	}

	headerHandler, err := block.GetHeaderFromBytes(notifier.marshaller, notifier.headerV2Creator, headerBytes)
	if err != nil {
		return nil, err
	}

	return headerHandler.(*block.HeaderV2), nil
}

func (notifier *sovereignNotifier) notifyHandlers(extendedHeader *block.ShardHeaderExtended) {
	notifier.mutHandler.RLock()
	defer notifier.mutHandler.RUnlock()

	for _, handler := range notifier.handlers {
		handler.SaveExtendedHeader(extendedHeader)
	}
}

// RegisterHandler will register an extended header handler to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHandler(handler process.ExtendedHeaderHandler) error {
	if handler == nil {
		return errNilExtendedHeaderHandler
	}

	notifier.mutHandler.Lock()
	notifier.handlers = append(notifier.handlers, handler)
	notifier.mutHandler.Unlock()

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
