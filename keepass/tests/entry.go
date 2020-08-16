package tests

import (
	"regexp"
	"testing"
	"time"

	k "github.com/mostfunkyduck/kp/keepass"
)

func RunTestNoParent(t *testing.T, r Resources) {
	name := "shmoo"
	e := r.Entry
	if !e.Set(k.Value{Name: "Title", Value: []byte(name)}) {
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
	expected += name
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

	expected = groupPath + newEntry.Title()
	if entryPath != expected {
		t.Fatalf("[%s] != [%s]", entryPath, expected)
	}
}

// kpv1 only supports a limited set of fields, so we have to let the caller
// specify what value to set

func RunTestEntryTimeFuncs(t *testing.T, r Resources) {
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
func RunTestEntryPasswordTitleFuncs(t *testing.T, r Resources) {
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

func RunTestOutput(t *testing.T, r Resources) {
	// only testing that it returns SOMETHING and doesn't bork
	if r.Entry.Output(true) == "" {
		t.Fatalf("output was empty")
	}
	if r.Entry.Output(false) == "" {
		t.Fatalf("output was empty")
	}
}

func RunTestSearchInNestedSubgroup(t *testing.T, r Resources) {
	sg, err := r.Group.NewSubgroup("RunTestSearchInNestedSubgroup")
	if err != nil {
		t.Fatalf(err.Error())
	}

	e, err := sg.NewEntry("askdfhjaskjfhasf")
	if err != nil {
		t.Fatalf(err.Error())
	}

	paths := r.Db.Root().Search(regexp.MustCompile(e.Title()))

	expected := "/" + r.Group.Name() + "/" + sg.Name() + "/" + e.Title()
	if paths[0] != expected {
		t.Fatalf("[%s] != [%s]", paths[0], expected)
	}
}
