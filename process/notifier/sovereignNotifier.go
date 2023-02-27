package notifier

import "github.com/multiversx/mx-chain-sovereign-notifier-go/config"

type sovereignNotifier struct {
	subscribedAddresses []string
}

func NewSovereignNotifier(config config.Config) (*sovereignNotifier, error) {
	if len(config.SubscribedAddresses) == 0 {
		return nil, errNoSubscribedAddresses
	}

	return &sovereignNotifier{
		subscribedAddresses: config.SubscribedAddresses,
	}, nil
}
