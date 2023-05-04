package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// HeaderSubscriberStub -
type HeaderSubscriberStub struct {
	AddHeaderCalled func(headerHash []byte, header sovereign.IncomingHeaderHandler)
}

// AddHeader -
func (stub *HeaderSubscriberStub) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) {
	if stub.AddHeaderCalled != nil {
		stub.AddHeaderCalled(headerHash, header)
	}
}

// IsInterfaceNil -
func (stub *HeaderSubscriberStub) IsInterfaceNil() bool {
	return stub == nil
}
