package factory

import (
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
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

	argsWsClient := &client.ArgsWsClient{
		OperationHandler:         operationHandler,
		Url:                      cfg.WebSocketConfig.Url,
		Uint64ByteSliceConverter: uint64ByteSlice.NewBigEndianConverter(),
	}
	return client.NewWsClient(argsWsClient)
}
