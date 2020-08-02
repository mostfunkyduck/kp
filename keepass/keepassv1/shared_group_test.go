package keepassv1_test

import (
	"testing"

	runner "github.com/mostfunkyduck/kp/keepass/tests"
)

func TestNestedSubGroupPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestNestedSubGroupPath(t, r)
}

func TestDoubleNestedGroupPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestDoubleNestedGroupPath(t, r)
}

func TestGroupParentFunctions(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestGroupParentFunctions(t, r)
}

func TestGroupUniqueness(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestGroupUniqueness(t, r)
}

func TestRemoveSubgroup(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestRemoveSubgroup(t, r)
}

func TestGroupEntryFuncs(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestGroupEntryFuncs(t, r)
}

func TestSubgroupSearch(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestSubgroupSearch(t, r)
}

func TestIsRoot(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestIsRoot(t, r)
}
