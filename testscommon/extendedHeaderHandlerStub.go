package testscommon

import "github.com/multiversx/mx-chain-core-go/data/block"

// ExtendedHeaderHandlerStub -
type ExtendedHeaderHandlerStub struct {
	ReceivedExtendedHeaderCalled func(header *block.ShardHeaderExtended)
}

// ReceivedExtendedHeader -
func (stub *ExtendedHeaderHandlerStub) ReceivedExtendedHeader(header *block.ShardHeaderExtended) {
	if stub.ReceivedExtendedHeaderCalled != nil {
		stub.ReceivedExtendedHeaderCalled(header)
	}
}

// IsInterfaceNil -
func (stub *ExtendedHeaderHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
