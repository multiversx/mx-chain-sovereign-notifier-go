package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

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

// SubscribedEvent contains a subscribed event from the main chain that the notifier is watching
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

// ArgsSovereignNotifier is a struct placeholder for args needed to create a sovereign notifier
type ArgsSovereignNotifier struct {
	Marshaller       marshal.Marshalizer
	Hasher           hashing.Hasher
	SubscribedEvents []SubscribedEvent
}

type sovereignNotifier struct {
	headersNotifier  *headersNotifier
	subscribedEvents []SubscribedEvent
	headerV2Creator  block.EmptyBlockCreator
	marshaller       marshal.Marshalizer
	hasher           hashing.Hasher
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(args ArgsSovereignNotifier) (*sovereignNotifier, error) {
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, errNilHasher
	}
	err := checkEvents(args.SubscribedEvents)
	if err != nil {
		return nil, err
	}

	return &sovereignNotifier{
		subscribedEvents: args.SubscribedEvents,
		headersNotifier:  newHeadersNotifier(),
		headerV2Creator:  block.NewEmptyHeaderV2Creator(),
		marshaller:       args.Marshaller,
		hasher:           args.Hasher,
	}, nil
}

func checkEvents(events []SubscribedEvent) error {
	if len(events) == 0 {
		return errNoSubscribedEvent
	}

	log.Debug("sovereign notifier received config", "num subscribed events", len(events))
	for idx, event := range events {
		if len(event.Identifier) == 0 {
			return fmt.Errorf("%w at event index = %d", errNoSubscribedIdentifier, idx)
		}

		log.Debug("sovereign notifier", "subscribed event identifier", string(event.Identifier))

		err := checkEmptyAddresses(event.Addresses)
		if err != nil {
			return fmt.Errorf("%w at event index = %d", err, idx)
		}
	}

	return nil
}

func checkEmptyAddresses(addresses map[string]string) error {
	if len(addresses) == 0 {
		return errNoSubscribedAddresses
	}

	for decodedAddr, encodedAddr := range addresses {
		if len(decodedAddr) == 0 || len(encodedAddr) == 0 {
			return errNoSubscribedAddresses
		}

		log.Debug("sovereign notifier", "decoded address - key", decodedAddr, "subscribed address - value", encodedAddr)
	}

	return nil
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
		IncomingEvents: notifier.createIncomingEventsFromTxData(outportBlock.TransactionPool.Transactions, outportBlock.TransactionPool.SmartContractResults),
	}

	headerHash, err := core.CalculateHash(notifier.marshaller, notifier.hasher, extendedHeader)
	if err != nil {
		return err
	}

	return notifier.headersNotifier.notifyHeaderSubscribers(extendedHeader, headerHash)
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
		eventsFromLog := notifier.getIncomingEvents(logData)
		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	return incomingEvents
}

func (notifier *sovereignNotifier) createIncomingEventsFromTxData(txs map[string]*outport.TxInfo, scrInfos map[string]*outport.SCRInfo) []*transaction.Event {
	incomingEvents := make([]*transaction.Event, 0)

	log.Info("this is from radu!!!!!")

	for _, tx := range txs {
		data := string(tx.Transaction.GetData())
		splitdata := strings.Split(data, "@")
		if len(splitdata) < 4 {
			continue
		}
		receiver := []byte(splitdata[1])
		log.Info("createIncomingEventsFromTxData", "txdata", data, "receiver", string(receiver))

		found := notifier.isSubscribedAddress(receiver)
		if !found {
			continue
		}
		eventsFromLog := notifier.addEvents(receiver, tx.Transaction.Data)
		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	for _, scrInfo := range scrInfos {
		receiver := scrInfo.SmartContractResult.RcvAddr
		data := string(scrInfo.SmartContractResult.GetData())
		log.Info("createIncomingEventsFromTxData", "scrdata", data)
		scr, _ := json.Marshal(scrInfo.SmartContractResult)
		log.Info("createIncomingEventsFromTxData", "scr", string(scr))
		found := notifier.isSubscribedAddress(receiver)
		if !found {
			continue
		}
		eventsFromLog := notifier.addEvents(scrInfo.SmartContractResult.RcvAddr, scrInfo.SmartContractResult.Data)
		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	return incomingEvents
}

func (notifier *sovereignNotifier) getIncomingEvents(logData *outport.LogData) []*transaction.Event {
	incomingEvents := make([]*transaction.Event, 0)

	for _, event := range logData.GetLog().Events {
		if !notifier.isSubscribed(event, logData.TxHash) {
			continue
		}

		incomingEvents = append(incomingEvents, event)
	}

	return incomingEvents
}

func (notifier *sovereignNotifier) addEvents(receiverAddress []byte, txData []byte) []*transaction.Event {
	log.Info("txData", "txData", string(txData))

	incomingEvents := make([]*transaction.Event, 0)

	transferESDT := [][]byte{
		[]byte("WEGLD-bd4d79"), // id
		big.NewInt(0).Bytes(),  // nonce = 0
		big.NewInt(50).Bytes(), // value
	}

	topic := append([][]byte{receiverAddress}, transferESDT...)

	event := &transaction.Event{
		Address:    receiverAddress,
		Identifier: []byte("deposit"),
		Topics:     topic,
		Data:       make([]byte, 0),
	}
	incomingEvents = append(incomingEvents, event)
	return incomingEvents
}

func (notifier *sovereignNotifier) isSubscribed(event *transaction.Event, txHash string) bool {
	for _, subEvent := range notifier.subscribedEvents {
		if !bytes.Equal(event.GetIdentifier(), subEvent.Identifier) {
			continue
		}

		receiver := event.GetAddress()
		encodedAddr, found := subEvent.Addresses[string(receiver)]
		if !found {
			continue
		}

		log.Info("found incoming event", "original tx hash", txHash, "receiver", encodedAddr)
		return true
	}

	return false
}

func (notifier *sovereignNotifier) isSubscribedAddress(receiver []byte) bool {
	receiverStr := string(receiver)
	for _, subEvent := range notifier.subscribedEvents {
		for address, encodedAddr := range subEvent.Addresses {
			log.Info("isSubscribedAddress", "address", address, "receiver", receiverStr, "encodedAddr", encodedAddr)
		}
		encodedAddr, found := subEvent.Addresses[receiverStr]
		if !found {
			log.Info("not found", "receiver", receiverStr)
			continue
		}

		log.Info("found incoming event", "receiver", encodedAddr)
		return true
	}

	return false
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

// RegisterHandler will register an extended header handler to be notified about incoming headers and miniblocks
func (notifier *sovereignNotifier) RegisterHandler(handler process.IncomingHeaderSubscriber) error {
	return notifier.headersNotifier.registerSubscriber(handler)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
