package tests

import (
	"regexp"
	"testing"
)

func RunTestNestedSubGroupPath(t *testing.T, r Resources) {
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

func RunTestDoubleNestedGroupPath(t *testing.T, r Resources) {
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

func RunTestPathOnOrphanedGroup(t *testing.T, r Resources) {
	if err := r.Db.Root().RemoveSubgroup(r.Group); err != nil {
		t.Fatalf(err.Error())
	}

	// if the path is obtained from root, there will be a preceding slash
	// otherwise, no slash
	if path, err := r.Group.Path(); path != r.Group.Name() {
		t.Fatalf("orphaned group somehow had a path: %s: %s", path, err)
	}

}

func RunTestGroupParentFunctions(t *testing.T, r Resources) {
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

func RunTestGroupUniqueness (t *testing.T, r Resources) {
	newGroupWrapper := r.BlankGroup
	newGroupWrapper.SetName(r.Entry.Title())

	if err := r.Group.AddSubgroup(newGroupWrapper); err == nil {
		t.Fatalf("added subgroup with same name as entry in group")
	}

	newGroupWrapper.SetName("asdf")
	if _, err := r.Group.NewSubgroup(newGroupWrapper.Name()); err != nil {
		t.Fatalf(err.Error())
	}
	if err := r.Group.AddSubgroup(newGroupWrapper); err == nil {
		t.Fatalf("added subgroup with same name as other subgroup in group")
	}
}

func RunTestRemoveSubgroup(t *testing.T, r Resources) {
	name := "TestRemoveSubgroup"

	originalLen := len(r.Group.Groups())
	sg, err := r.Group.NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(r.Group.Groups()) != originalLen + 1 {
		t.Fatalf("[%d] != [%d]", len(r.Group.Groups()), originalLen + 1)
	}
	if err := r.Group.RemoveSubgroup(sg); err != nil {
		t.Fatalf(err.Error())
	}

	if len(r.Group.Groups()) != originalLen {
		t.Fatalf("[%d] != [%d]", len(r.Group.Groups()), originalLen)
	}

	if err := r.Group.RemoveSubgroup(sg); err == nil {
		t.Fatalf("removed subgroup twice")
	}
}

func RunTestGroupEntryFuncs (t *testing.T, r Resources) {
	if err := r.Group.AddEntry(r.Entry); err == nil {
		t.Fatalf("added duplicate entry: [%v][%v]", r.Entry, r.Group)
	}

	originalLen := len(r.Group.Entries())
	if err := r.Group.RemoveEntry(r.Entry); err != nil {
		t.Fatalf(err.Error())
	}

	if len(r.Group.Entries()) != originalLen - 1 {
		t.Fatalf("[%d] != [%d]", len(r.Group.Entries()), originalLen -1)
	}

	if err := r.Group.RemoveEntry(r.Entry); err == nil {
		t.Fatalf("successfully removed non existent entry")
	}
}

func RunTestSubgroupSearch(t *testing.T, r Resources) {
	name := "TestSubgroupSearch"
	sg, err := r.Group.NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	paths := r.Group.Search(regexp.MustCompile(sg.Name()))
	if len(paths) != 1 {
		t.Fatalf("too many search results")
	}

	sgPath, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	sgPath += "/" // subgroups always get at erminal slash in paths
	if paths[0] != sgPath {
		t.Fatalf("[%s] != [%s]", paths[0], sgPath)
	}
}

func RunTestIsRoot(t *testing.T, r Resources) {
	if r.Group.IsRoot() {
		t.Fatalf("non root group thinks it's root")
	}

	newGroupWrapper := r.BlankGroup
	if newGroupWrapper.IsRoot() {
		t.Fatalf("orphaned group with no parent thinks it's root")
	}
}

