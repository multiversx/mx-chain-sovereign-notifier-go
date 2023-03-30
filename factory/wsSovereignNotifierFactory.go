package factory

import (
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/indexer"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
)

// CreatWsSovereignNotifier will create a ws sovereign shard notifier
func CreatWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	sovereignNotifier, err := notifier.NewSovereignNotifier(cfg)
	if err != nil {
		return nil, err
	}

	cache, err := indexer.NewLRUOutportBlockCache(cfg.WebSocketConfig.BlocksCacheSize)
	if err != nil {
		return nil, err
	}
	dataIndexer, err := indexer.NewIndexer(sovereignNotifier, cache)
	if err != nil {
		return nil, err
	}

	marshaller, err := factory.NewMarshalizer(cfg.WebSocketConfig.MarshallerType)
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
