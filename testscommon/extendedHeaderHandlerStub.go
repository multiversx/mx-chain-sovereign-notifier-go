package testscommon

import "github.com/multiversx/mx-chain-core-go/data/block"

// ExtendedHeaderHandlerStub -
type ExtendedHeaderHandlerStub struct {
	SaveExtendedHeaderCalled func(header *block.ShardHeaderExtended)
}

// SaveExtendedHeader -
func (stub *ExtendedHeaderHandlerStub) SaveExtendedHeader(header *block.ShardHeaderExtended) {
	if stub.SaveExtendedHeaderCalled != nil {
		stub.SaveExtendedHeaderCalled(header)
	}
}
