package testscommon

// TransactionSubscriberStub -
type TransactionSubscriberStub struct {
	AddDataCalled func(key []byte, data interface{}, sizeInBytes int, cacheId string)
}

// AddData -
func (tss *TransactionSubscriberStub) AddData(key []byte, data interface{}, sizeInBytes int, cacheId string) {
	if tss.AddDataCalled != nil {
		tss.AddDataCalled(key, data, sizeInBytes, cacheId)
	}
}

// IsInterfaceNil -
func (tss *TransactionSubscriberStub) IsInterfaceNil() bool {
	return tss == nil
}
