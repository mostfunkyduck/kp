package main_test

import (
	"strings"
	"testing"

	main "github.com/mostfunkyduck/kp"
)

func TestMkdir(t *testing.T) {
	// Happy path
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
	l, e, err := main.TraversePath(r.Db, r.Db.CurrentLocation(), r.Db.CurrentLocation().Path()+groupName)
	if err != nil {
		t.Fatalf("could not traverse path: %s", err)
	}

	if e != nil {
		t.Fatalf("entry found instead of target for new group\n")
	}

	expected := r.Db.CurrentLocation().Path() + groupName + "/"
	if l.Path() != expected {
		t.Fatalf("[%s] != [%s]", l.Path(), expected)
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
