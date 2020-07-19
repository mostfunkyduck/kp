package keepassv2_test

import (
	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/v2"
	g "github.com/tobischo/gokeepasslib/v3"
	"testing"
)

func TestPathNoParent(t *testing.T) {
	name := "shmoo"
	db := main.NewDatabase(g.NewDatabase(), "/dev/null", k.Options{})
	newEnt := g.NewEntry()
	e := main.WrapEntry(&newEnt, db)
	if !e.Set(k.Value{Name: "Title", Value: name}) {
		t.Fatalf("could not set title")
	}
	output, err := e.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// this guy has no parent, shouldn't even have the root "/" in the path
	if output != name {
		t.Fatalf("[%s] !+ [%s]", output, name)
	}
}

func TestRegularPath(t *testing.T) {
	name := "asldkfjalskdfjasldkfjasfd"
	r := createTestResources(t)
	e, err := r.Group.NewEntry(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	path, err := e.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected += "/" + name
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}
}
