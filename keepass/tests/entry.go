package tests

import (
	"testing"
	"time"
	k "github.com/mostfunkyduck/kp/keepass"
)

type Resources struct {
	Db k.Database
	Entry k.Entry
	Group k.Group
	// BlankEntry and BlankGroup are empty resources for testing freshly
	// allocated structs
	BlankEntry k.Entry
	BlankGroup k.Group
}

func RunTestNoParent(t *testing.T, r Resources) {
	name := "shmoo"
	e := r.Entry
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

	if parent := e.Parent(); parent != nil {
		t.Fatalf("%v", parent)
	}
}

func RunTestRegularPath(t *testing.T, r Resources) {
	name := "asldkfjalskdfjasldkfjasfd"
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

	parent := r.Entry.Parent()
	if parent == nil {
		t.Fatalf("%v", r)
	}

	parentPath, err := parent.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	groupPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if parentPath != groupPath {
		t.Fatalf("[%s] != [%s]", parentPath, groupPath)
	}


	newEntry := r.BlankEntry
	if err := newEntry.SetParent(r.Group); err != nil {
		t.Fatalf(err.Error())
	}

	entryPath, err := newEntry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	groupPath, err = r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected = groupPath + "/" + newEntry.Title()
	if entryPath != expected {
		t.Fatalf("[%s] != [%s]", entryPath, expected)
	}
}

func RunTestGetSet(t *testing.T, r Resources) {
	value := k.Value {
		Name: "TestEntryGetSet",
		Value: "test value",
	}

	if r.Entry.Get(value.Name) != (k.Value{}) {
		t.Fatalf("initial get should have returned empty value")
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

func RunTestEntryTimeFuncs (t *testing.T, r Resources) {
	newTime := time.Now().Add(time.Duration(1) * time.Hour)
	r.Entry.SetCreationTime(newTime)
	if !r.Entry.CreationTime().Equal(newTime) {
		t.Fatalf("%v, %v", newTime, r.Entry.CreationTime())
	}

	newTime = newTime.Add(time.Duration(1) * time.Hour)
	r.Entry.SetLastModificationTime(newTime)
	if !r.Entry.LastModificationTime().Equal(newTime) {
		t.Fatalf("%v, %v", newTime, r.Entry.LastModificationTime())
	}

	newTime = newTime.Add(time.Duration(1) * time.Hour)
	r.Entry.SetLastAccessTime(newTime)
	if !r.Entry.LastAccessTime().Equal(newTime) {
		t.Fatalf("%v, %v", newTime, r.Entry.LastAccessTime())
	}
}
func RunTestEntryPasswordTitleFuncs (t *testing.T, r Resources) {
	password := "swordfish"
	r.Entry.SetPassword(password)
	if r.Entry.Password() != password {
		t.Fatalf("[%s] != [%s]", r.Entry.Password(), password)
	}

	title := "blobulence"
	r.Entry.SetTitle(title)
	if r.Entry.Title() != title {
		t.Fatalf("[%s] != [%s]", r.Entry.Title(), title)
	}
}

func RunTestOutput (t *testing.T, r Resources) {
	// only testing that it returns SOMETHING and doesn't bork
	if r.Entry.Output(true) == "" {
		t.Fatalf("output was empty")
	}
	if r.Entry.Output(false) == "" {
		t.Fatalf("output was empty")
	}
}
