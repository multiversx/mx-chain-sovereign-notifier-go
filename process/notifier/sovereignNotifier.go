package notifier

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
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

// Notify will notify the sovereign nodes via p2p about the finalized block and incoming mb txs
func (notifier *sovereignNotifier) Notify(_ *outport.OutportBlock) error {
	return nil
}

func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
