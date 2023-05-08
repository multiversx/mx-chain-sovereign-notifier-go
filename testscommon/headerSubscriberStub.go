package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// HeaderSubscriberStub -
type HeaderSubscriberStub struct {
	AddHeaderCalled func(headerHash []byte, header sovereign.IncomingHeaderHandler) error
}

// AddHeader -
func (stub *HeaderSubscriberStub) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	if stub.AddHeaderCalled != nil {
		return stub.AddHeaderCalled(headerHash, header)
	}

	return nil
}

// IsInterfaceNil -
func (stub *HeaderSubscriberStub) IsInterfaceNil() bool {
	return stub == nil
}
