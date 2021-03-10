package commands_test

import (
	"fmt"
	"strings"
	"testing"

	main "github.com/mostfunkyduck/kp/internal/commands"
)

func createGroup(group string, r testResources) error {
	r.Context.Args = []string{
		group,
	}
	if _, err := r.Readline.WriteStdin([]byte("n")); err != nil {
		return fmt.Errorf("could not write to readline: %s", err)
	}

	// FIXME the v2 tests  won't pass without this repetiton
	if _, err := r.Readline.WriteStdin([]byte("n")); err != nil {
		return fmt.Errorf("could not write to readline: %s", err)
	}
	main.NewGroup(r.Shell)(r.Context)
	return nil
}

func verifyGroup(group string, r testResources) error {
	currentLocation := r.Db.CurrentLocation()
	l, e, err := main.TraversePath(r.Db, currentLocation, group)
	if err != nil {
		return fmt.Errorf("could not traverse path from [%s] to [%s]: %s", currentLocation.Name(), group, err)
	}

	if e != nil {
		return fmt.Errorf("entry found instead of target for new group")
	}

	path, err := r.Db.CurrentLocation().Path()
	if err != nil {
		return fmt.Errorf("could not locate path of current DB location: %s", err)
	}
	expected := path + group + "/"
	lPath, err := l.Path()
	if err != nil {
		return fmt.Errorf("could not locate location: %s", err)
	}
	if lPath != expected {
		return fmt.Errorf("[%s] != [%s]", lPath, expected)
	}
	return nil
}

func TestMkdir(t *testing.T) {
	// Happy path, testing the first group and the second will, in effect, test nested groups
	r := createTestResources(t)
	r.Db.SetCurrentLocation(r.Group)
	if err := createGroup("test2", r); err != nil {
		t.Fatal(err)
	}

	if err := verifyGroup("test2", r); err != nil {
		t.Fatal(err)
	}

}

func TestMkdirTerminalSlash(t *testing.T) {
	// Happy path, testing the first group and the second will, in effect, test nested groups
	r := createTestResources(t)
	r.Db.SetCurrentLocation(r.Group)
	if err := createGroup("test2/", r); err != nil {
		t.Fatal(err)
	}

	if err := verifyGroup("test2", r); err != nil {
		t.Fatal(err)
	}
}

func TestMkdirNestedSubgroup(t *testing.T) {
	// Happy path
	r := createTestResources(t)
	r.Db.SetCurrentLocation(r.Group)
	if err := createGroup("test2", r); err != nil {
		t.Fatalf("could not create group: %s\n", err)
	}
	if err := verifyGroup("test2", r); err != nil {
		t.Fatalf("could not verify group: %s\n", err)
	}

	r.Db.SetCurrentLocation(r.Group.Groups()[0])
	if err := createGroup("test3", r); err != nil {
		t.Fatalf("could not create nested group: %s\n", err)
	}
	if err := verifyGroup("test3", r); err != nil {
		t.Fatalf("could not verify nested group: %s\n", err)
	}
}

func TestMkdirGroupNameIdenticalToEntry(t *testing.T) {
	r := createTestResources(t)
	r.Db.SetCurrentLocation(r.Group)

	if err := createGroup(r.Entry.Title(), r); err != nil {
		t.Fatalf("could not create nested group: %s\n", err)
	}
}

func TestMkdirGroupNameDuplicate(t *testing.T) {
	r := createTestResources(t)
	r.F.outputHolder.output = ""
	if err := createGroup(r.Group.Name(), r); err != nil {
		t.Fatal(err)
	}

	o := r.F.outputHolder.output
	expected := "cannot create duplicate"
	if !strings.Contains(o, expected) {
		t.Fatalf("[%s] does not contain [%s]", o, expected)
	}
}
