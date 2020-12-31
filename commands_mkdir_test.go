package main_test

import (
	"strings"
	"testing"

	main "github.com/mostfunkyduck/kp"
)

func TestMkdir(t *testing.T) {
	// Happy path, testing the first group and the second will, in effect, test nested groups
	r := createTestResources(t)
	groupName := "test2"
	r.Db.SetCurrentLocation(r.Group)
	r.Context.Args = []string{
		groupName,
	}
	if _, err := r.Readline.WriteStdin([]byte("n")); err != nil {
		t.Fatalf(err.Error())
	}
	main.NewGroup(r.Shell)(r.Context)
	path, err := r.Db.CurrentLocation().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	l, e, err := main.TraversePath(r.Db, r.Db.CurrentLocation(), path+groupName)
	if err != nil {
		t.Fatalf("could not traverse path: %s", err)
	}

	if e != nil {
		t.Fatalf("entry found instead of target for new group\n")
	}

	path, err = r.Db.CurrentLocation().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := path + groupName + "/"
	lPath, err := l.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if lPath != expected {
		t.Fatalf("[%s] != [%s]", lPath, expected)
	}

	r.F.outputHolder.output = ""
	// Testing a duplicate while we're at it
	main.NewGroup(r.Shell)(r.Context)
	o := r.F.outputHolder.output
	expected = "cannot create duplicate"
	if !strings.Contains(o, expected) {
		t.Fatalf("[%s] does not contain [%s]", o, expected)
	}
}

func TestMkdirNestedSubgroup(t *testing.T) {
	// Happy path, testing the first group and the second will, in effect, test nested groups
	r := createTestResources(t)
	groupName := "test2/"
	r.Db.SetCurrentLocation(r.Group)
	r.Context.Args = []string{
		groupName,
	}
	if _, err := r.Readline.WriteStdin([]byte("n")); err != nil {
		t.Fatalf(err.Error())
	}
	main.NewGroup(r.Shell)(r.Context)

	r.Db.SetCurrentLocation(r.Group.Groups()[0])
	groupName2 := "test3"
	r.Context.Args = []string{
		groupName2,
	}
	if _, err := r.Readline.WriteStdin([]byte("n")); err != nil {
		t.Fatalf(err.Error())
	}
	main.NewGroup(r.Shell)(r.Context)

	l, e, err := main.TraversePath(r.Db, r.Db.CurrentLocation(), groupName2)
	if err != nil {
		t.Fatalf("could not traverse path: %s", err)
	}

	if e != nil {
		t.Fatalf("entry found instead of target for new group\n")
	}

	path, err := r.Db.CurrentLocation().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := path + groupName2 + "/"
	lPath, err := l.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if lPath != expected {
		t.Fatalf("[%s] != [%s]", lPath, expected)
	}

	r.F.outputHolder.output = ""
	// Testing a duplicate while we're at it
	main.NewGroup(r.Shell)(r.Context)
	o := r.F.outputHolder.output
	expected = "cannot create duplicate"
	if !strings.Contains(o, expected) {
		t.Fatalf("[%s] does not contain [%s]", o, expected)
	}
}
