package notifier

import (
	sovCore "github.com/multiversx/mx-chain-core-go/core/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// IncomingHeadersNotifierHandler defines an incoming header notifier behavior
type IncomingHeadersNotifierHandler interface {
	NotifyHeaderSubscribers(header sovereign.IncomingHeaderHandler) error
	RegisterSubscriber(handler sovCore.IncomingHeaderSubscriber) error
}
