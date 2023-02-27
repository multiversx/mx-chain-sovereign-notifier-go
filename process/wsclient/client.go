package wsclient

import "github.com/multiversx/mx-chain-sovereign-notifier-go/process"

type client struct {
	notifier process.SovereignNotifier
}

func NewWsClient(notifier process.SovereignNotifier) (*client, error) {
	if notifier == nil {
		return nil, errNilSovereignNotifier
	}

	return &client{
		notifier: notifier,
	}, nil
}

func (c *client) Start() {

}

func (c *client) Close() {

}
