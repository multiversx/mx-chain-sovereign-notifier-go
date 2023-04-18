package testscommon

// ShardCoordinatorStub -
type ShardCoordinatorStub struct {
	ComputeIdCalled func(address []byte) uint32
}

// ComputeId -
func (scs *ShardCoordinatorStub) ComputeId(address []byte) uint32 {
	if scs.ComputeIdCalled != nil {
		return scs.ComputeIdCalled(address)
	}
	return 0
}

// IsInterfaceNil -
func (scs *ShardCoordinatorStub) IsInterfaceNil() bool {
	return scs == nil
}
