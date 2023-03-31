package testscommon

import "github.com/multiversx/mx-chain-core-go/data/outport"

// OutportBlockCacheStub -
type OutportBlockCacheStub struct {
	AddCalled     func(outportBlock *outport.OutportBlock) error
	ExtractCalled func(headerHash []byte) (*outport.OutportBlock, error)
}

// Add -
func (obc *OutportBlockCacheStub) Add(outportBlock *outport.OutportBlock) error {
	if obc.AddCalled != nil {
		return obc.AddCalled(outportBlock)
	}
	return nil
}

// Extract -
func (obc *OutportBlockCacheStub) Extract(headerHash []byte) (*outport.OutportBlock, error) {
	if obc.ExtractCalled != nil {
		return obc.ExtractCalled(headerHash)
	}

	return nil, nil
}

// IsInterfaceNil -
func (obc *OutportBlockCacheStub) IsInterfaceNil() bool {
	return obc == nil
}
