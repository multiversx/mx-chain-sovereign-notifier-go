package wsclient

import "github.com/multiversx/mx-chain-sovereign-notifier-go/process"

type client struct {
	notifier process.SovereignNotifier
}

// NewWsClient will create a ws client to receive data from an observer/light client
func NewWsClient(notifier process.SovereignNotifier) (*client, error) {
	if notifier == nil {
		return nil, errNilSovereignNotifier
	}

	return &client{
		notifier: notifier,
	}, nil
}

// Start will start the client listening ws process
func (c *client) Start() {

}

// Close will close the underlying ws connection
func (c *client) Close() {

}
