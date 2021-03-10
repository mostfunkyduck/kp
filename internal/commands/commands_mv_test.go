package commands_test

import (
	"testing"

	main "github.com/mostfunkyduck/kp/internal/commands"
)

func TestMv(t *testing.T) {
	r := createTestResources(t)
	newName := "example"
	rGroupPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	rootPath, err := r.Db.Root().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{
		rGroupPath,
		rootPath + newName,
	}
	main.Mv(r.Shell)(r.Context)
	rGroupPath, err = r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	rootPath, err = r.Db.Root().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if rGroupPath != rootPath+newName+"/" {
		t.Fatalf("[%s] != [%s]", rGroupPath, rootPath+newName+"/")
	}
}

// Verify that you can't overwrite a group with a group
func TestMvGroupOverwriteGroup(t *testing.T) {
	r := createTestResources(t)
	originalGroupCount := len(r.Db.Root().Groups())
	g, _ := r.Group.NewSubgroup(r.Group.Name())
	r.Db.SetCurrentLocation(r.Db.Root())
	originalGroupPath, err := g.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	rootPath, err := r.Db.Root().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{
		originalGroupPath,
		rootPath,
	}
	main.Mv(r.Shell)(r.Context)
	// test that the group didn't get moved
	gPath, err := g.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if gPath != originalGroupPath {
		t.Fatalf("[%s] != [%s] (%s)", gPath, originalGroupPath, r.F.outputHolder.output)
	}

	// make sure it didn't add a spurious third group during a botched move
	if len(r.Db.Root().Groups()) != originalGroupCount {
		t.Fatalf("[%d] != [%d] (%s)", len(r.Db.Root().Groups()), originalGroupCount, r.F.outputHolder.output)
	}
}

func TestMvGroupOverwriteEntry(t *testing.T) {
	r := createTestResources(t)
	originalGroupPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	originalEntryPath, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{
		originalGroupPath,
		originalEntryPath,
	}
	main.Mv(r.Shell)(r.Context)
	groupPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test that the group didn't get moved
	if groupPath != originalGroupPath {
		t.Fatalf("[%s] != [%s]", groupPath, originalGroupPath)
	}

	entryPath, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if entryPath != originalEntryPath {
		t.Fatalf("[%s] != [%s]", entryPath, originalEntryPath)
	}
}

func TestMvEntryIntoGroup(t *testing.T) {
	r := createTestResources(t)

	newName := "test2"
	g, _ := r.Db.Root().NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	originalEntryPath, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	gPath, err := g.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{
		originalEntryPath,
		gPath,
	}
	main.Mv(r.Shell)(r.Context)

	gPath, err = g.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedPath := gPath + r.Entry.Title()

	entryPath, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if entryPath != expectedPath {
		t.Fatalf("[%s] != [%s]", entryPath, expectedPath)
	}
}

func TestMvGroupIntoGroup(t *testing.T) {
	r := createTestResources(t)
	newName := "test"
	g, _ := r.Group.NewSubgroup(newName)
	r.Db.SetCurrentLocation(r.Db.Root())

	gPath, err := g.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	rGrpPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	r.Context.Args = []string{
		gPath,
		rGrpPath,
	}

	// NOTE this was broken beforehand
	main.Mv(r.Shell)(r.Context)
	// testing that the group is now a subgroup
	expectedPath := rGrpPath + g.Name() + "/"
	if gPath != expectedPath {
		t.Fatalf("[%s] != [%s] (%s)", gPath, expectedPath, r.F.outputHolder.output)
	}

}
