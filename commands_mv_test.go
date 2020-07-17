package main_test

import (
	main "github.com/mostfunkyduck/kp"
	"testing"
)

func TestMv(t *testing.T) {
	r := createTestResources(t)
	newName := "example"
	r.Context.Args = []string{
		r.Group.Path(),
		r.Db.Root().Path() + newName,
	}
	main.Mv(r.Shell)(r.Context)
	if r.Group.Path() != r.Db.Root().Path()+newName+"/" {
		t.Fatalf("[%s] != [%s]", r.Group.Path(), r.Db.Root().Path()+newName+"/")
	}
}

// Verify that you can't overwrite a group with a group
func TestMvGroupOverwriteGroup(t *testing.T) {
	r := createTestResources(t)
	originalGroupCount := len(r.Db.Root().Groups())
	g, _ := r.Group.NewSubgroup(r.Group.Name())
	r.Db.SetCurrentLocation(r.Db.Root())
	originalGroupPath := g.Path()
	r.Context.Args = []string{
		originalGroupPath,
		r.Db.Root().Path(),
	}
	main.Mv(r.Shell)(r.Context)
	// test that the group didn't get moved
	if g.Path() != originalGroupPath {
		t.Fatalf("[%s] != [%s] (%s)", g.Path(), originalGroupPath, r.F.outputHolder.output)
	}

	// make sure it didn't add a spurious third group during a botched move
	if len(r.Db.Root().Groups()) != originalGroupCount {
		t.Fatalf("[%d] != [%d] (%s)", len(r.Db.Root().Groups()), originalGroupCount, r.F.outputHolder.output)
	}
}

func TestMvGroupOverwriteEntry(t *testing.T) {
	r := createTestResources(t)
	originalGroupPath := r.Group.Path()
	originalEntryPath := r.Entry.Path()
	r.Context.Args = []string{
		originalGroupPath,
		r.Entry.Path(),
	}
	main.Mv(r.Shell)(r.Context)
	// test that the group didn't get moved
	if r.Group.Path() != originalGroupPath {
		t.Fatalf("[%s] != [%s]", r.Group.Path(), originalGroupPath)
	}

	if r.Entry.Path() != originalEntryPath {
		t.Fatalf("[%s] != [%s]", r.Entry.Path(), originalEntryPath)
	}
}

func TestMvEntryIntoGroup(t *testing.T) {
	r := createTestResources(t)

	newName := "test2"
	g, _ := r.Db.Root().NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	originalEntryPath := r.Entry.Path()
	r.Context.Args = []string{
		originalEntryPath,
		g.Path(),
	}
	main.Mv(r.Shell)(r.Context)

	expectedPath := g.Path() + r.Entry.Get("title").Value.(string)
	if r.Entry.Path() != expectedPath {
		t.Fatalf("[%s] != [%s]", r.Entry.Path(), expectedPath)
	}
}

func TestMvGroupIntoGroup(t *testing.T) {
	r := createTestResources(t)
	newName := "test"
	g, _ := r.Group.NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	r.Context.Args = []string{
		g.Path(),
		r.Group.Path(),
	}

	// testing that the group is now a subgroup
	expectedPath := r.Group.Path() + g.Name() + "/"
	if g.Path() != expectedPath {
		t.Fatalf("[%s] != [%s] (%s)", g.Path(), expectedPath, r.F.outputHolder.output)
	}

}
