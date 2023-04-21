package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// HeaderSubscriberStub -
type HeaderSubscriberStub struct {
	AddHeaderCalled func(headerHash []byte, header data.HeaderHandler)
}

// AddHeader -
func (stub *HeaderSubscriberStub) AddHeader(headerHash []byte, header data.HeaderHandler) {
	if stub.AddHeaderCalled != nil {
		stub.AddHeaderCalled(headerHash, header)
	}
}

// IsInterfaceNil -
func (stub *HeaderSubscriberStub) IsInterfaceNil() bool {
	return stub == nil
}
