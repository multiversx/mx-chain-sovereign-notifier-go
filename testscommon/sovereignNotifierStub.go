package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

// SovereignNotifierStub -
type SovereignNotifierStub struct {
	NotifyCalled          func(finalizedBlock *outport.OutportBlock) error
	RegisterHandlerCalled func(handler process.ExtendedHeaderHandler) error
}

// Notify -
func (sn *SovereignNotifierStub) Notify(finalizedBlock *outport.OutportBlock) error {
	if sn.NotifyCalled != nil {
		return sn.NotifyCalled(finalizedBlock)
	}

	return nil
}

// RegisterHandler -
func (sn *SovereignNotifierStub) RegisterHandler(handler process.ExtendedHeaderHandler) error {
	if sn.RegisterHandlerCalled != nil {
		return sn.RegisterHandlerCalled(handler)
	}

	return nil
}

// IsInterfaceNil -
func (sn *SovereignNotifierStub) IsInterfaceNil() bool {
	return sn == nil
}
