package keepassv1_test

import (
	"regexp"
	"testing"

	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func TestTitle(t *testing.T) {
	title := "test"
	e := &keepass.Entry{Title: title}
	wrapper := v1.WrapEntry(e, &v1.Database{})
	wrapperTitle := wrapper.Title()
	if wrapperTitle != title {
		t.Fatalf("%s != %s", title, wrapperTitle)
	}
}

func TestEntrySearch(t *testing.T) {
	title := "TestEntrySearch"
	db, err := keepass.New(&keepass.Options{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	dbWrapper := v1.NewDatabase(db, "/dev/null")
	sg, err := dbWrapper.Root().NewSubgroup("DOESN'T MATCH")
	if err != nil {
		t.Fatal(err)
	}

	wrapper, err := sg.NewEntry(title)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := wrapper.Search(regexp.MustCompile("TestEntry.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("%v", paths)
	}

	path, err := wrapper.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if paths[0] != path {
		t.Fatalf("[%s] != [%s]", paths[0], path)
	}
}
