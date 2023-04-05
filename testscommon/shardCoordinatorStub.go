package testscommon

// ShardCoordinatorStub -
type ShardCoordinatorStub struct {
}

// ComputeId -
func (scs *ShardCoordinatorStub) ComputeId(_ []byte) uint32 {
	return 0
}

// IsInterfaceNil -
func (scs *ShardCoordinatorStub) IsInterfaceNil() bool {
	return scs == nil
}
