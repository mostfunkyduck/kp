package main_test

import (
	"strings"
	"testing"
	main "github.com/mostfunkyduck/kp"
)

func TestMkdir(t *testing.T) {
	// Happy path
	r := createTestResources(t)
	groupName := "test"
	r.Context.Args = []string{
		groupName,
	}
	main.NewGroup(r.Shell)(r.Context)
	l, err := r.Db.TraversePath(r.Db.CurrentLocation(), r.Db.CurrentLocation().Pwd() + groupName)
	if err != nil {
		t.Fatalf("could not traverse path: %s", err)
	}

	expected := r.Db.CurrentLocation().Pwd() + groupName + "/" 
	if l.Pwd() != expected {
		t.Fatalf("[%s] != [%s]", l.Pwd(), expected)
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
