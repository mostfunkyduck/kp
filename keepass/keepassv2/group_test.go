package keepassv2_test

import (
	"testing"
)

func TestNestedSubGroupPath(t *testing.T) {
	r := createTestResources(t)
	sgName := "blipblip"
	sg, err := r.Group.NewSubgroup(sgName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	path, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "/" + r.Group.Name() + "/" + sgName
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}
}

func TestDoubleNestedGroupPath(t *testing.T) {
	r := createTestResources(t)
	sgName := "blipblip"
	sg, err := r.Group.NewSubgroup(sgName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	sg1, err := sg.NewSubgroup(sgName + "1")
	if err != nil {
		t.Fatalf(err.Error())
	}

	sgPath, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	sg1Path, err := sg1.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	sgExpected := "/" + r.Group.Name() + "/" + sgName
	if sgPath != sgExpected {
		t.Fatalf("[%s] != [%s]", sgPath, sgExpected)
	}

	sg1Expected := sgExpected + "/" + sgName + "1"
	if sg1Path != sg1Expected {
		t.Fatalf("[%s] != [%s]", sgPath, sgExpected)
	}
}

func TestPathOnOrphanedGroup(t *testing.T) {
	r := createTestResources(t)
	if err := r.Db.Root().RemoveSubgroup(r.Group); err != nil {
		t.Fatalf(err.Error())
	}

	// if the path is obtained from root, there will be a preceding slash
	// otherwise, no slash
	if path, err := r.Group.Path(); path != r.Group.Name() {
		t.Fatalf("orphaned group somehow had a path: %s: %s", path, err)
	}

}

func TestGroupParentFunctions(t *testing.T) {
	r := createTestResources(t)
	name := "TestGroupParentFunctions"

	// first test 'parent' when it is returning the root group
	sg, err := r.Db.Root().NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	parent := sg.Parent()
	if !parent.IsRoot() {
		t.Fatalf("subgroup of root group was not pointing at root")
	}

	// now  test when parent returns a regular group
	subsg, err := sg.NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	parentUUID, err := subsg.Parent().UUIDString()
	if err != nil {
		t.Fatalf(err.Error())
	}
	sgUUID, err := sg.UUIDString()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if sgUUID != parentUUID {
		t.Fatalf("[%s] != [%s]", sgUUID, parentUUID)
	}
}
