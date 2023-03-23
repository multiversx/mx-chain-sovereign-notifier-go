package testscommon

import "github.com/multiversx/mx-chain-core-go/data/outport"

type IndexerStub struct {
	SaveBlockCalled      func(outportBlock *outport.OutportBlock) error
	FinalizedBlockCalled func(finalizedBlock *outport.FinalizedBlock) error
}

func (is *IndexerStub) SaveBlock(outportBlock *outport.OutportBlock) error {
	if is.SaveBlockCalled != nil {
		return is.SaveBlockCalled(outportBlock)
	}

	return nil
}

func (is *IndexerStub) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	if is.FinalizedBlockCalled != nil {
		return is.FinalizedBlockCalled(finalizedBlock)
	}

	return nil
}

func (is *IndexerStub) IsInterfaceNil() bool {
	return is == nil
}
