package notifier

import (
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

type sovereignNotifier struct {
	mutHandler          sync.RWMutex
	handlers            []process.ExtendedHeaderHandler
	subscribedAddresses map[string]struct{}
	headerV2Creator     block.EmptyBlockCreator
	marshaller          marshal.Marshalizer
	shardCoordinator    process.ShardCoordinator
}

// ArgsSovereignNotifier is a struct placeholder for args needed to create a sovereign notifier
type ArgsSovereignNotifier struct {
	ShardCoordinator    process.ShardCoordinator
	Marshaller          marshal.Marshalizer
	SubscribedAddresses [][]byte
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(args ArgsSovereignNotifier) (*sovereignNotifier, error) {
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, errNilShardCoordinator
	}

	subscribedAddresses, err := getAddressesMap(args.SubscribedAddresses)
	if err != nil {
		return nil, err
	}

	log.Debug("received config", "subscribed addresses", args.SubscribedAddresses)

	return &sovereignNotifier{
		subscribedAddresses: subscribedAddresses,
		handlers:            make([]process.ExtendedHeaderHandler, 0),
		mutHandler:          sync.RWMutex{},
		headerV2Creator:     block.NewEmptyHeaderV2Creator(),
		marshaller:          args.Marshaller,
		shardCoordinator:    args.ShardCoordinator,
	}, nil
}

func getAddressesMap(addresses [][]byte) (map[string]struct{}, error) {
	numAddresses := len(addresses)
	if numAddresses == 0 {
		return nil, errNoSubscribedAddresses
	}

	addressesMap := make(map[string]struct{}, numAddresses)
	for _, addr := range addresses {
		addressesMap[string(addr)] = struct{}{}
	}

	if len(addressesMap) != numAddresses {
		return nil, errDuplicateSubscribedAddresses
	}

	return addressesMap, nil
}

// Notify will notify the sovereign nodes about the finalized block and incoming mb txs
// For each subscribed address, it searches if the receiver is found in transaction pool.
// If found, IncomingMiniBlocks will contain the ordered tx hashes by execution.
func (notifier *sovereignNotifier) Notify(outportBlock *outport.OutportBlock) error {
	err := checkNilOutportBlockFields(outportBlock)
	if err != nil {
		return err
	}

	mbs, err := notifier.createAllIncomingMbs(outportBlock.TransactionPool)
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
		return errNilOutportBlock
	}
	if outportBlock.TransactionPool == nil {
		return errNilTransactionPool
	}
	if outportBlock.BlockData == nil {
		return errNilBlockData
	}

	return nil
}

func (notifier *sovereignNotifier) createAllIncomingMbs(txPool *outport.TransactionPool) ([]*block.MiniBlock, error) {
	mbs := make([]*block.MiniBlock, 0)

	txsMb, err := notifier.createIncomingMbsFromTxs(txPool.Transactions)
	if err != nil {
		return nil, err
	}

	// TODO: when specs are defined, we should also handle scrs mbs
	mbs = append(mbs, txsMb...)
	return mbs, nil
}

func (notifier *sovereignNotifier) createIncomingMbsFromTxs(txs map[string]*outport.TxInfo) ([]*block.MiniBlock, error) {
	execOrderTxHashMap := make(map[string]uint32)
	shardIDTxHashMap := make(map[uint32][][]byte)

	for txHash, tx := range txs {
		receiver := tx.GetTransaction().GetRcvAddr()
		_, found := notifier.subscribedAddresses[string(receiver)]
		if !found {
			continue
		}

		hashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, fmt.Errorf("%w, hash: %s", err, txHash)
		}

		log.Info("found incoming tx", "tx hash", txHash, "receiver", hex.EncodeToString(receiver))

		execOrderTxHashMap[string(hashBytes)] = tx.GetExecutionOrder()

		sender := tx.GetTransaction().GetSndAddr()
		senderShardID := notifier.shardCoordinator.ComputeId(sender)
		shardIDTxHashMap[senderShardID] = append(shardIDTxHashMap[senderShardID], hashBytes)
	}

	mbs := make([]*block.MiniBlock, 0, len(shardIDTxHashMap))
	for shardID, txHashesInShard := range shardIDTxHashMap {
		sort.SliceStable(txHashesInShard, func(i, j int) bool {
			return execOrderTxHashMap[string(txHashesInShard[i])] < execOrderTxHashMap[string(txHashesInShard[j])]
		})

		mbs = append(mbs, &block.MiniBlock{
			TxHashes:        txHashesInShard,
			ReceiverShardID: core.SovereignChainShardId,
			SenderShardID:   shardID,
			Type:            block.TxBlock,
			Reserved:        nil,
		})
	}

	sort.SliceStable(mbs, func(i, j int) bool {
		return mbs[i].SenderShardID < mbs[j].SenderShardID
	})
	return mbs, nil
}

func (notifier *sovereignNotifier) getHeaderV2(headerType core.HeaderType, headerBytes []byte) (*block.HeaderV2, error) {
	if headerType != core.ShardHeaderV2 {
		return nil, fmt.Errorf("%w : %s, expected: %s",
			errInvalidHeaderTypeReceived, headerType, core.ShardHeaderV2)
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
		handler.ReceivedExtendedHeader(extendedHeader)
	}
}

// RegisterHandler will register an extended header handler to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHandler(handler process.ExtendedHeaderHandler) error {
	if check.IfNil(handler) {
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
