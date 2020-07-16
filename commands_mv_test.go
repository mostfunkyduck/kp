package main_test

import (
	main "github.com/mostfunkyduck/kp"
	"testing"
)

func TestMv(t *testing.T) {
	r := createTestResources(t)
	newName := "example"
	r.Context.Args = []string{
		r.Group.Pwd(),
		r.Db.Root().Pwd() + newName,
	}
	main.Mv(r.Shell)(r.Context)
	if r.Group.Pwd() != r.Db.Root().Pwd()+newName+"/" {
		t.Fatalf("[%s] != [%s]", r.Group.Pwd(), r.Db.Root().Pwd()+newName+"/")
	}
}

// Verify that you can't overwrite a group with a group
func TestMvGroupOverwriteGroup(t *testing.T) {
	r := createTestResources(t)
	originalGroupCount := len(r.Db.Root().Groups())
	g := r.Group.NewSubgroup(r.Group.Name())
	r.Db.SetCurrentLocation(r.Db.Root())
	originalGroupPath := g.Pwd()
	r.Context.Args = []string{
		originalGroupPath,
		r.Db.Root().Pwd(),
	}
	main.Mv(r.Shell)(r.Context)
	// test that the group didn't get moved
	if g.Pwd() != originalGroupPath {
		t.Fatalf("[%s] != [%s] (%s)", g.Pwd(), originalGroupPath, r.F.outputHolder.output)
	}

	// make sure it didn't add a spurious third group during a botched move
	if len(r.Db.Root().Groups()) != originalGroupCount {
		t.Fatalf("[%d] != [%d] (%s)", len(r.Db.Root().Groups()), originalGroupCount, r.F.outputHolder.output)
	}
}

func TestMvGroupOverwriteEntry(t *testing.T) {
	r := createTestResources(t)
	originalGroupPwd := r.Group.Pwd()
	originalEntryPwd := r.Entry.Pwd()
	r.Context.Args = []string{
		originalGroupPwd,
		r.Entry.Pwd(),
	}
	main.Mv(r.Shell)(r.Context)
	// test that the group didn't get moved
	if r.Group.Pwd() != originalGroupPwd {
		t.Fatalf("[%s] != [%s]", r.Group.Pwd(), originalGroupPwd)
	}

	if r.Entry.Pwd() != originalEntryPwd {
		t.Fatalf("[%s] != [%s]", r.Entry.Pwd(), originalEntryPwd)
	}
}

func TestMvEntryIntoGroup(t *testing.T) {
	r := createTestResources(t)

	newName := "test2"
	g := r.Db.Root().NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	originalEntryPwd := r.Entry.Pwd()
	r.Context.Args = []string{
		originalEntryPwd,
		g.Pwd(),
	}
	main.Mv(r.Shell)(r.Context)

	expectedPath := g.Pwd() + r.Entry.Get("title").Value.(string)
	if r.Entry.Pwd() != expectedPath {
		t.Fatalf("[%s] != [%s]", r.Entry.Pwd(), expectedPath)
	}
}

func TestMvGroupIntoGroup(t *testing.T) {
	r := createTestResources(t)
	newName := "test"
	g := r.Group.NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	r.Context.Args = []string{
		g.Pwd(),
		r.Group.Pwd(),
	}

	// testing that the group is now a subgroup
	expectedPath := r.Group.Pwd() + g.Name() + "/"
	if g.Pwd() != expectedPath {
		t.Fatalf("[%s] != [%s] (%s)", g.Pwd(), expectedPath, r.F.outputHolder.output)
	}

}
