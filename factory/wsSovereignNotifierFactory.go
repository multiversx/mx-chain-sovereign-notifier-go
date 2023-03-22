package factory

import (
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/indexer"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/wsclient"
)

// CreatWsSovereignNotifier will create a ws sovereign shard notifier
func CreatWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	sovereignNotifier, err := notifier.NewSovereignNotifier(cfg)
	if err != nil {
		return nil, err
	}

	dataIndexer, err := indexer.NewIndexer(sovereignNotifier)
	if err != nil {
		return nil, err
	}

	marshaller, err := factory.NewMarshalizer(cfg.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}

	operationHandler, err := indexer.NewOperationHandler(dataIndexer, marshaller)
	if err != nil {
		return nil, err
	}

	return wsclient.NewWsClient(operationHandler)
}
