package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

// SovereignNotifierStub -
type SovereignNotifierStub struct {
	NotifyCalled                   func(finalizedBlock *outport.OutportBlock) error
	RegisterHeaderSubscriberCalled func(handler process.HeaderSubscriber) error
}

// Notify -
func (sn *SovereignNotifierStub) Notify(finalizedBlock *outport.OutportBlock) error {
	if sn.NotifyCalled != nil {
		return sn.NotifyCalled(finalizedBlock)
	}

	return nil
}

// RegisterHeaderSubscriber -
func (sn *SovereignNotifierStub) RegisterHeaderSubscriber(handler process.HeaderSubscriber) error {
	if sn.RegisterHeaderSubscriberCalled != nil {
		return sn.RegisterHeaderSubscriberCalled(handler)
	}

	return nil
}

func (sn *SovereignNotifierStub) RegisterTxSubscriber(_ process.TransactionSubscriber) error {
	return nil
}

// IsInterfaceNil -
func (sn *SovereignNotifierStub) IsInterfaceNil() bool {
	return sn == nil
}
