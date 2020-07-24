package keepassv2_test

import (
	"testing"

	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/keepassv2"
	g "github.com/tobischo/gokeepasslib/v3"
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

func TestEntrySet (t *testing.T) {
	r := createTestResources(t)
	value := k.Value {
		Name: "TestEntrySet",
		Value: "test value",
	}
	if !r.Entry.Set(value) {
		t.Fatalf("could not set value")
	}

	entryValue := r.Entry.Get(value.Name).Value.(string)
	if entryValue != value.Value {
		t.Fatalf("[%s] != [%s], %v", entryValue, value.Name, value)
	}

	secondValue := "asldkfj"
	value.Value = secondValue
	if !r.Entry.Set(value) {
		t.Fatalf("could not overwrite value: %v", value)
	}

	entryValue = r.Entry.Get(value.Name).Value.(string)
	if entryValue != secondValue {
		t.Fatalf("[%s] != [%s] %v", entryValue, secondValue, value)
	}
}
