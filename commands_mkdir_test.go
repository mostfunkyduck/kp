package main_test

import (
	"testing"
	main "github.com/mostfunkyduck/kp"
)

func TestMkdir(t *testing.T) {
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
}
