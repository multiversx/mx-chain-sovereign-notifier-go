package factory

import (
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/notifier"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process/wsclient"
)

// CreatWsSovereignNotifier will create a ws sovereign shard notifier
func CreatWsSovereignNotifier(cfg config.Config) (process.WSClient, error) {
	sovereignNotifier, err := notifier.NewSovereignNotifier(cfg)
	if err != nil {
		return nil, err
	}

	return wsclient.NewWsClient(sovereignNotifier)
}
