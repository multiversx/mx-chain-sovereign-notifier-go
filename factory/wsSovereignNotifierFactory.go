package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	hashingFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/indexer"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
)

const (
	addressLen = 32
)

// ArgsCreateSovereignNotifier is a struct placeholder for sovereign notifier args
type ArgsCreateSovereignNotifier struct {
	MarshallerType   string
	HasherType       string
	SubscribedEvents []config.SubscribedEvent
}

// CreateSovereignNotifier creates a sovereign notifier which will notify subscribed handlers about incoming headers
func CreateSovereignNotifier(args ArgsCreateSovereignNotifier) (process.SovereignNotifier, error) {
	marshaller, err := factory.NewMarshalizer(args.MarshallerType)
	if err != nil {
		return nil, err
	}

	hasher, err := hashingFactory.NewHasher(args.HasherType)
	if err != nil {
		return nil, err
	}

	subscribedEvents, err := getSubscribedEvents(args.SubscribedEvents)
	if err != nil {
		return nil, err
	}

	argsSovereignNotifier := notifier.ArgsSovereignNotifier{
		Marshaller:       marshaller,
		Hasher:           hasher,
		SubscribedEvents: subscribedEvents,
	}
	return notifier.NewSovereignNotifier(argsSovereignNotifier)
}

// ArgsWsClientReceiverNotifier is a struct placeholder for ws client receiver args
type ArgsWsClientReceiverNotifier struct {
	WebSocketConfig   config.WebSocketConfig
	SovereignNotifier process.SovereignNotifier
}

// CreateWsClientReceiverNotifier creates a ws client receiver for incoming outport blocks
func CreateWsClientReceiverNotifier(args ArgsWsClientReceiverNotifier) (process.WSClient, error) {
	marshaller, err := factory.NewMarshalizer(args.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}

	cache := indexer.NewOutportBlockCache()
	dataIndexer, err := indexer.NewIndexer(args.SovereignNotifier, cache)
	if err != nil {
		return nil, err
	}

	payloadProcessor, err := indexer.NewPayloadProcessor(dataIndexer, marshaller)
	if err != nil {
		return nil, err
	}

	argsWsClient := client.ArgsCreateWsClient{
		Url:                args.WebSocketConfig.Url,
		RetryDurationInSec: args.WebSocketConfig.RetryDuration,
		BlockingAckOnError: args.WebSocketConfig.BlockingAckOnError,
		PayloadProcessor:   payloadProcessor,
	}
	return client.CreateWsClient(argsWsClient)
}

// CreateWsSovereignNotifier will create a ws sovereign shard notifier
func CreateWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	sovereignNotifier, err := CreateSovereignNotifier(ArgsCreateSovereignNotifier{
		MarshallerType:   cfg.WebSocketConfig.MarshallerType,
		SubscribedEvents: cfg.SubscribedEvents,
		HasherType:       cfg.WebSocketConfig.HasherType,
	})
	if err != nil {
		return nil, err
	}

	return CreateWsClientReceiverNotifier(ArgsWsClientReceiverNotifier{
		WebSocketConfig:   cfg.WebSocketConfig,
		SovereignNotifier: sovereignNotifier,
	})
}

func getSubscribedEvents(events []config.SubscribedEvent) ([]notifier.SubscribedEvent, error) {
	ret := make([]notifier.SubscribedEvent, len(events))
	for idx, event := range events {
		addressesMap, err := getAddressesMap(event.Addresses)
		if err != nil {
			return nil, fmt.Errorf("%w for event at index = %d", err, idx)
		}

		ret[idx] = notifier.SubscribedEvent{
			Identifier: []byte(event.Identifier),
			Addresses:  addressesMap,
		}
	}

	return ret, nil
}

func getAddressesMap(addresses []string) (map[string]string, error) {
	numAddresses := len(addresses)
	if numAddresses == 0 {
		return nil, errNoSubscribedAddresses
	}

	pubKeyConv, err := pubkeyConverter.NewBech32PubkeyConverter(addressLen, core.DefaultAddressPrefix)
	if err != nil {
		return nil, err
	}

	addressesMap := make(map[string]string, numAddresses)
	for _, encodedAddr := range addresses {
		decodedAddr, errDecode := pubKeyConv.Decode(encodedAddr)
		if errDecode != nil {
			return nil, errDecode
		}

		addressesMap[string(decodedAddr)] = encodedAddr
	}

	if len(addressesMap) != numAddresses {
		return nil, errDuplicateSubscribedAddresses
	}

	return addressesMap, nil
}
