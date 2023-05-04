package notifier

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type sovereignNotifier struct {
	mutHandler          sync.RWMutex
	handlers            []process.IncomingHeaderSubscriber
	subscribedAddresses map[string]struct{}
	headerV2Creator     block.EmptyBlockCreator
	marshaller          marshal.Marshalizer
	hasher              hashing.Hasher
}

// ArgsSovereignNotifier is a struct placeholder for args needed to create a sovereign notifier
type ArgsSovereignNotifier struct {
	Marshaller          marshal.Marshalizer
	Hasher              hashing.Hasher
	SubscribedAddresses [][]byte
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(args ArgsSovereignNotifier) (*sovereignNotifier, error) {
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
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
		handlers:            make([]process.IncomingHeaderSubscriber, 0),
		mutHandler:          sync.RWMutex{},
		headerV2Creator:     block.NewEmptyHeaderV2Creator(),
		marshaller:          args.Marshaller,
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

	headerV2, err := notifier.getHeaderV2(core.HeaderType(outportBlock.BlockData.HeaderType), outportBlock.BlockData.HeaderBytes)
	if err != nil {
		return err
	}

	extendedHeader := &sovereign.IncomingHeader{
		Header:         headerV2,
		IncomingEvents: notifier.createIncomingEvents(outportBlock.TransactionPool.Logs),
	}

	return notifier.notifyHandlers(extendedHeader)
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

func (notifier *sovereignNotifier) createIncomingEvents(logsData []*outport.LogData) []*transaction.Event {
	incomingEvents := make([]*transaction.Event, 0)

	for _, logData := range logsData {
		events := logData.GetLog().Events
		eventsFromLog := notifier.getIncomingEvents(events, logData.TxHash)

		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	return incomingEvents
}

func (notifier *sovereignNotifier) getIncomingEvents(events []*transaction.Event, txHash string) []*transaction.Event {
	incomingEvents := make([]*transaction.Event, 0)

	for _, event := range events {
		receiver := event.GetAddress()
		_, found := notifier.subscribedAddresses[string(receiver)]
		if !found {
			continue
		}

		if !(bytes.Compare(event.GetIdentifier(), []byte("deposit")) == 0) {
			continue
		}

		log.Info("found incoming event", "original tx hash", txHash, "receiver", hex.EncodeToString(receiver))
		incomingEvents = append(incomingEvents, event)
	}

	return incomingEvents
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

func (notifier *sovereignNotifier) notifyHandlers(header sovereign.IncomingHeaderHandler) error {
	headerHash, err := core.CalculateHash(notifier.marshaller, notifier.hasher, header)
	if err != nil {
		return err
	}

	log.Info("notifying shard extended header", "hash", hex.EncodeToString(headerHash))

	notifier.mutHandler.RLock()
	defer notifier.mutHandler.RUnlock()

	for _, handler := range notifier.handlers {
		handler.AddHeader(headerHash, header)
	}

	return nil
}

// RegisterHandler will register an extended header handler to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHandler(handler process.IncomingHeaderSubscriber) error {
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
