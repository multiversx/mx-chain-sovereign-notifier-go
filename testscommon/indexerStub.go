package testscommon

import "github.com/multiversx/mx-chain-core-go/data/outport"

// IndexerStub -
type IndexerStub struct {
	SaveBlockCalled      func(outportBlock *outport.OutportBlock) error
	FinalizedBlockCalled func(finalizedBlock *outport.FinalizedBlock) error
}

// SaveBlock -
func (is *IndexerStub) SaveBlock(outportBlock *outport.OutportBlock) error {
	if is.SaveBlockCalled != nil {
		return is.SaveBlockCalled(outportBlock)
	}

	return nil
}

// FinalizedBlock -
func (is *IndexerStub) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	if is.FinalizedBlockCalled != nil {
		return is.FinalizedBlockCalled(finalizedBlock)
	}

	return nil
}

// IsInterfaceNil -
func (is *IndexerStub) IsInterfaceNil() bool {
	return is == nil
}
