package notifier

import (
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	shardProcess "github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type sovereignNotifier struct {
	mutHeaderHandlers   sync.RWMutex
	mutTxHandlers       sync.RWMutex
	headerSubscribers   []process.HeaderSubscriber
	txSubscribers       []process.TransactionSubscriber
	subscribedAddresses map[string]struct{}
	headerV2Creator     block.EmptyBlockCreator
	marshaller          marshal.Marshalizer
	shardCoordinator    process.ShardCoordinator
	hasher              hashing.Hasher
}

// ArgsSovereignNotifier is a struct placeholder for args needed to create a sovereign notifier
type ArgsSovereignNotifier struct {
	ShardCoordinator    process.ShardCoordinator
	Marshaller          marshal.Marshalizer
	Hasher              hashing.Hasher
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
	if check.IfNil(args.Hasher) {
		return nil, errNilHasher
	}

	subscribedAddresses, err := getAddressesMap(args.SubscribedAddresses)
	if err != nil {
		return nil, err
	}

	log.Debug("received config", "subscribed addresses", args.SubscribedAddresses)

	return &sovereignNotifier{
		subscribedAddresses: subscribedAddresses,
		headerSubscribers:   make([]process.HeaderSubscriber, 0),
		txSubscribers:       make([]process.TransactionSubscriber, 0),
		mutHeaderHandlers:   sync.RWMutex{},
		mutTxHandlers:       sync.RWMutex{},
		headerV2Creator:     block.NewEmptyHeaderV2Creator(),
		marshaller:          args.Marshaller,
		shardCoordinator:    args.ShardCoordinator,
		hasher:              args.Hasher,
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

	return notifier.notifyHeaderSubscribers(extendedHeader)
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

	txsMb, _, err := notifier.createIncomingMbsFromTxs(txPool.Transactions)
	if err != nil {
		return nil, err
	}

	// TODO: when specs are defined, we should also handle scrs mbs
	mbs = append(mbs, txsMb...)
	return mbs, nil
}

func (notifier *sovereignNotifier) createIncomingMbsFromTxs(txs map[string]*outport.TxInfo) ([]*block.MiniBlock, map[string]*transaction.Transaction, error) {
	execOrderTxHashMap := make(map[string]uint32)
	shardIDTxHashMap := make(map[uint32][][]byte)
	incomingTxs := make(map[string]*transaction.Transaction)

	for txHash, tx := range txs {
		receiver := tx.GetTransaction().GetRcvAddr()
		_, found := notifier.subscribedAddresses[string(receiver)]
		if !found {
			continue
		}

		hashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, nil, fmt.Errorf("%w, hash: %s", err, txHash)
		}

		log.Info("found incoming tx", "tx hash", txHash, "receiver", hex.EncodeToString(receiver))

		execOrderTxHashMap[string(hashBytes)] = tx.GetExecutionOrder()

		sender := tx.GetTransaction().GetSndAddr()
		senderShardID := notifier.shardCoordinator.ComputeId(sender)
		shardIDTxHashMap[senderShardID] = append(shardIDTxHashMap[senderShardID], hashBytes)

		incomingTxs[string(hashBytes)] = tx.GetTransaction()
	}

	return createSortedMbs(execOrderTxHashMap, shardIDTxHashMap, block.TxBlock), incomingTxs, nil
}

func createSortedMbs(
	execOrderTxHashMap map[string]uint32,
	shardIDTxHashMap map[uint32][][]byte,
	mbType block.Type,
) []*block.MiniBlock {
	mbs := make([]*block.MiniBlock, 0, len(shardIDTxHashMap))

	for shardID, txHashesInShard := range shardIDTxHashMap {
		sortTxs(txHashesInShard, execOrderTxHashMap)

		mbs = append(mbs, &block.MiniBlock{
			TxHashes:        txHashesInShard,
			ReceiverShardID: core.SovereignChainShardId,
			SenderShardID:   shardID,
			Type:            mbType,
			Reserved:        nil,
		})
	}

	sort.SliceStable(mbs, func(i, j int) bool {
		return mbs[i].SenderShardID < mbs[j].SenderShardID
	})

	return mbs
}

func sortTxs(txs [][]byte, execOrderTxHashMap map[string]uint32) {
	sort.SliceStable(txs, func(i, j int) bool {
		return execOrderTxHashMap[string(txs[i])] < execOrderTxHashMap[string(txs[j])]
	})
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

func (notifier *sovereignNotifier) notifyHeaderSubscribers(header data.HeaderHandler) error {
	headerHash, err := core.CalculateHash(notifier.marshaller, notifier.hasher, header)
	if err != nil {
		return err
	}

	log.Info("notifying shard extended header", "hash", hex.EncodeToString(headerHash))

	notifier.mutHeaderHandlers.RLock()
	defer notifier.mutHeaderHandlers.RUnlock()

	for _, handler := range notifier.headerSubscribers {
		handler.AddHeader(headerHash, header)
	}

	return nil
}

func (notifier *sovereignNotifier) notifyTxSubscribers(txs map[string]*transaction.Transaction) {
	notifier.mutTxHandlers.RLock()
	defer notifier.mutTxHandlers.RUnlock()

	for _, subscriber := range notifier.txSubscribers {
		notifyTxSubscriber(subscriber, txs)
	}
}

func notifyTxSubscriber(subscriber process.TransactionSubscriber, txs map[string]*transaction.Transaction) {
	for txHashStr, tx := range txs {
		cacheID := shardProcess.ShardCacherIdentifier(0, core.SovereignChainShardId) // TODO: Decided here senderShardID
		subscriber.AddData([]byte(txHashStr), tx, tx.Size(), cacheID)
	}
}

// RegisterHandler will register an extended header handler to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHandler(handler process.HeaderSubscriber) error {
	if check.IfNil(handler) {
		return errNilExtendedHeaderHandler
	}

	notifier.mutHeaderHandlers.Lock()
	notifier.headerSubscribers = append(notifier.headerSubscribers, handler)
	notifier.mutHeaderHandlers.Unlock()

	return nil
}

// RegisterTxSubscriber will register a tx subscriber to be notified about incoming transactions
func (notifier *sovereignNotifier) RegisterTxSubscriber(handler process.TransactionSubscriber) error {
	if check.IfNil(handler) {
		return errNilExtendedHeaderHandler
	}

	notifier.mutTxHandlers.Lock()
	notifier.txSubscribers = append(notifier.txSubscribers, handler)
	notifier.mutTxHandlers.Unlock()

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
