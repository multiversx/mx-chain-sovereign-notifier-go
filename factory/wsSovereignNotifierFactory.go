package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/indexer"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
)

const addressLen = 32

// CreatWsSovereignNotifier will create a ws sovereign shard notifier
func CreatWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	marshaller, err := factory.NewMarshalizer(cfg.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}

	subscribedAddresses, err := getDecodedAddresses(cfg.SubscribedAddresses)
	if err != nil {
		return nil, err
	}

	argsSovereignNotifier := notifier.ArgsSovereignNotifier{
		Marshaller:          marshaller,
		SubscribedAddresses: subscribedAddresses,
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
