package notifier

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type txHandlerInfo struct {
	tx            data.TransactionHandler
	senderShardID uint32
	hash          []byte
}

type sovereignNotifier struct {
	headerNotifier      *headersNotifier
	txsNotifier         *txsNotifier
	subscribedAddresses map[string]struct{}
	headerV2Creator     block.EmptyBlockCreator
	marshaller          marshal.Marshalizer
	shardCoordinator    process.ShardCoordinator
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
		headerNotifier:      newHeadersNotifier(args.Marshaller, args.Hasher),
		txsNotifier:         newTxsNotifier(),
		subscribedAddresses: subscribedAddresses,
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

	mbs, txs, err := notifier.createAllIncomingMbsAndTxs(outportBlock.TransactionPool)
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

	err = notifier.headerNotifier.notifyHeaderSubscribers(extendedHeader)
	if err != nil {
		return err
	}

	notifier.txsNotifier.notifyTxSubscribers(txs)
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

func (notifier *sovereignNotifier) createAllIncomingMbsAndTxs(txPool *outport.TransactionPool) ([]*block.MiniBlock, []*txHandlerInfo, error) {
	mbs := make([]*block.MiniBlock, 0)
	txs := make([]*txHandlerInfo, 0)

	txMbs, incomingTxs, err := notifier.createIncomingMbsAndTxs(txPool.Transactions)
	if err != nil {
		return nil, nil, err
	}
	txs = append(txs, createTxHandlersInfo(incomingTxs, txMbs)...)

	// TODO: when specs are defined, we should also handle scrs mbs
	mbs = append(mbs, txMbs...)
	return mbs, txs, nil
}

func (notifier *sovereignNotifier) createIncomingMbsAndTxs(txs map[string]*outport.TxInfo) ([]*block.MiniBlock, map[string]data.TransactionHandler, error) {
	execOrderTxHashMap := make(map[string]uint32)
	shardIDTxHashMap := make(map[uint32][][]byte)
	incomingTxs := make(map[string]data.TransactionHandler)

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

		incomingTxs[string(hashBytes)] = tx.GetTransaction()
		execOrderTxHashMap[string(hashBytes)] = tx.GetExecutionOrder()

		sender := tx.GetTransaction().GetSndAddr()
		senderShardID := notifier.shardCoordinator.ComputeId(sender)
		shardIDTxHashMap[senderShardID] = append(shardIDTxHashMap[senderShardID], hashBytes)
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
			SenderShardID:   core.MainChainShardId,
			Type:            mbType,
			Reserved:        []byte(fmt.Sprintf("%d", shardID)),
		})
	}

	sort.SliceStable(mbs, func(i, j int) bool {
		return bytes.Compare(mbs[i].Reserved, mbs[j].Reserved) <= 0
	})

	return mbs
}

func createTxHandlersInfo(txs map[string]data.TransactionHandler, mbs []*block.MiniBlock) []*txHandlerInfo {
	ret := make([]*txHandlerInfo, 0, len(txs))

	for _, mb := range mbs {
		for _, txHash := range mb.TxHashes {
			ret = append(ret, &txHandlerInfo{
				tx:            txs[string(txHash)],
				senderShardID: mb.SenderShardID,
				hash:          txHash,
			})
		}
	}

	return ret
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

// RegisterHeaderSubscriber will register a header subscriber to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHeaderSubscriber(handler process.HeaderSubscriber) error {
	return notifier.headerNotifier.registerSubscriber(handler)
}

// RegisterTxSubscriber will register a tx subscriber to be notified about incoming transactions
func (notifier *sovereignNotifier) RegisterTxSubscriber(handler process.TransactionSubscriber) error {
	return notifier.txsNotifier.registerSubscriber(handler)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
