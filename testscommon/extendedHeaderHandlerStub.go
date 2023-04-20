package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExtendedHeaderHandlerStub -
type ExtendedHeaderHandlerStub struct {
	AddHeaderCalled func(headerHash []byte, header data.HeaderHandler)
}

// AddHeader -
func (stub *ExtendedHeaderHandlerStub) AddHeader(headerHash []byte, header data.HeaderHandler) {
	if stub.AddHeaderCalled != nil {
		stub.AddHeaderCalled(headerHash, header)
	}
}

// IsInterfaceNil -
func (stub *ExtendedHeaderHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
