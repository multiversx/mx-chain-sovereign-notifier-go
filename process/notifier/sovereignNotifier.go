package notifier

import (
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type sovereignNotifier struct {
	subscribedAddresses []string
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(config config.Config) (*sovereignNotifier, error) {
	addresses := config.SubscribedAddresses
	if len(addresses) == 0 {
		return nil, errNoSubscribedAddresses
	}

	log.Debug("received config", "subscribed addresses", addresses)

	return &sovereignNotifier{
		subscribedAddresses: addresses,
	}, nil
}
