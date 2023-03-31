package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/indexer"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
)

// CreatWsSovereignNotifier will create a ws sovereign shard notifier
func CreatWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	container, err := createBlockCreatorsContainer()
	if err != nil {
		return nil, err
	}

	marshaller, err := factory.NewMarshalizer(cfg.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}

	argsSovereignNotifier := notifier.ArgsSovereignNotifier{
		Marshaller:          marshaller,
		BlockContainer:      container,
		SubscribedAddresses: nil,
	}
	sovereignNotifier, err := notifier.NewSovereignNotifier(argsSovereignNotifier)
	if err != nil {
		return nil, err
	}

	cache := indexer.NewOutportBlockCache()
	dataIndexer, err := indexer.NewIndexer(sovereignNotifier, cache)
	if err != nil {
		return nil, err
	}

	payloadProcessor, err := indexer.NewPayloadProcessor(dataIndexer, marshaller)
	if err != nil {
		return nil, err
	}

	argsWsClient := client.ArgsCreateWsClient{
		Url:                cfg.WebSocketConfig.Url,
		RetryDurationInSec: cfg.WebSocketConfig.RetryDuration,
		BlockingAckOnError: cfg.WebSocketConfig.BlockingAckOnError,
		PayloadProcessor:   payloadProcessor,
	}

	return client.CreateWsClient(argsWsClient)
}

func createBlockCreatorsContainer() (process.BlockContainerHandler, error) {
	container := block.NewEmptyBlockCreatorsContainer()
	err := container.Add(core.ShardHeaderV2, block.NewEmptyHeaderV2Creator())
	if err != nil {
		return nil, err
	}

	return container, nil
}
