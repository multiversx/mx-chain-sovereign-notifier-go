package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
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
	selfId     = 0
)

// ArgsCreateSovereignNotifier is a struct placeholder for sovereign notifier args
type ArgsCreateSovereignNotifier struct {
	MarshallerType      string
	HasherType          string
	SubscribedAddresses []string
	NumOfMainShards     uint32
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

	subscribedAddresses, err := getDecodedAddresses(args.SubscribedAddresses)
	if err != nil {
		return nil, err
	}

	shardCoordinator, err := sharding.NewMultiShardCoordinator(args.NumOfMainShards, selfId)
	if err != nil {
		return nil, err
	}

	argsSovereignNotifier := notifier.ArgsSovereignNotifier{
		Marshaller:          marshaller,
		Hasher:              hasher,
		SubscribedAddresses: subscribedAddresses,
		ShardCoordinator:    shardCoordinator,
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
		MarshallerType:      cfg.WebSocketConfig.MarshallerType,
		SubscribedAddresses: cfg.SubscribedAddresses,
		NumOfMainShards:     cfg.NumOfMainShards,
		HasherType:          cfg.WebSocketConfig.HasherType,
	})
	if err != nil {
		return nil, err
	}

	return CreateWsClientReceiverNotifier(ArgsWsClientReceiverNotifier{
		WebSocketConfig:   cfg.WebSocketConfig,
		SovereignNotifier: sovereignNotifier,
	})
}

func getDecodedAddresses(addresses []string) ([][]byte, error) {
	pubKeyConv, err := pubkeyConverter.NewBech32PubkeyConverter(addressLen, core.DefaultAddressPrefix)
	if err != nil {
		return nil, err
	}

	ret := make([][]byte, len(addresses))
	for idx, addr := range addresses {
		decodedAddr, errDecode := pubKeyConv.Decode(addr)
		if errDecode != nil {
			return nil, errDecode
		}

		ret[idx] = decodedAddr
	}

	return ret, err
}
