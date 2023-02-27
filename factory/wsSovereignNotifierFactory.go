package factory

import (
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/wsclient"
)

func CreatSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	sovereignNotifier, err := notifier.NewSovereignNotifier(cfg)
	if err != nil {
		return nil, err
	}

	return wsclient.NewWsClient(sovereignNotifier)
}
