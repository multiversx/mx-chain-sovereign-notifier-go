package wsclient

import "github.com/multiversx/mx-chain-sovereign-notifier-go/process"

type client struct {
	operationHandler process.OperationHandler
}

// NewWsClient will create a ws client to receive data from an observer/light client
func NewWsClient(operationHandler process.OperationHandler) (*client, error) {
	if operationHandler == nil {
		return nil, errNilOperationHandler
	}

	return &client{
		operationHandler: operationHandler,
	}, nil
}

// Start will start the client listening ws process
func (c *client) Start() {

}

// Close will close the underlying ws connection
func (c *client) Close() {

}
