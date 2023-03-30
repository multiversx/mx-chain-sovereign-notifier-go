package testscommon

import "github.com/multiversx/mx-chain-core-go/data/outport"

type SovereignNotifierStub struct {
	NotifyCalled func(finalizedBlock *outport.OutportBlock) error
}

func (sn *SovereignNotifierStub) Notify(finalizedBlock *outport.OutportBlock) error {
	if sn.NotifyCalled != nil {
		return sn.NotifyCalled(finalizedBlock)
	}

	return nil
}

func (sn *SovereignNotifierStub) IsInterfaceNil() bool {
	return sn == nil
}
