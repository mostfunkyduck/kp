package keepassv2_test

import (
	"testing"
	"time"

	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/keepassv2"
	g "github.com/tobischo/gokeepasslib/v3"
)

func TestNoParent(t *testing.T) {
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

	if parent := e.Parent(); parent != nil {
		t.Fatalf("%v", parent)
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


	looseEntry := g.NewEntry()
	newEntry := main.WrapEntry(&looseEntry, r.Db)
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

func TestEntryGetSet (t *testing.T) {
	r := createTestResources(t)

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

func TestEntryTimeFuncs (t *testing.T) {
	r := createTestResources(t)
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

func TestEntryPasswordTitleFuncs (t *testing.T) {
	r := createTestResources(t)
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

func TestOutput (t *testing.T) {
	r := createTestResources(t)
	// only testing that it returns SOMETHING and doesn't bork
	if r.Entry.Output(true) == "" {
		t.Fatalf("output was empty")
	}
	if r.Entry.Output(false) == "" {
		t.Fatalf("output was empty")
	}
}
