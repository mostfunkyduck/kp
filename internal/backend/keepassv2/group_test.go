package keepassv2_test

import (
	"testing"

	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
)

func TestNestedSubGroupPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestNestedSubGroupPath(t, r)
}

func TestDoubleNestedGroupPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestDoubleNestedGroupPath(t, r)
}

func TestPathOnOrphanedGroup(t *testing.T) {
	r := createTestResources(t)
	if err := r.Db.Root().RemoveSubgroup(r.Group); err != nil {
		t.Fatalf(err.Error())
	}

	// if the path is obtained from root, there will be a preceding slash
	// otherwise, no slash
	if path, err := r.Group.Path(); path != r.Group.Name()+"/" {
		t.Fatalf("orphaned group somehow had a path: %s: %s", path, err)
	}

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
