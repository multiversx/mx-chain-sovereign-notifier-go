package testscommon

// MarshallerStub -
type MarshallerStub struct {
	MarshalCalled func(obj interface{}) ([]byte, error)
}

// Marshal -
func (mm *MarshallerStub) Marshal(obj interface{}) ([]byte, error) {
	if mm.MarshalCalled != nil {
		return mm.MarshalCalled(obj)
	}

	return nil, nil
}

// Unmarshal -
func (mm *MarshallerStub) Unmarshal(_ interface{}, _ []byte) error {
	return nil
}

// IsInterfaceNil -
func (mm *MarshallerStub) IsInterfaceNil() bool {
	return mm == nil
}
