package main_test

import (
	"regexp"
	"testing"
)

func TestSearchFullPath(t *testing.T) {
	r := createTestResources(t)
	term, err := regexp.Compile(r.Entry.Title())
	if err != nil {
		t.Fatal(err)
	}
	paths, err := r.Group.Search(term)
	if err != nil {
		t.Fatal(err)
	}
	// the group and entry should match
	if len(paths) != 2 {
		t.Fatalf("%d != %d", len(paths), 1)
	}

	path, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if paths[1] != path {
		t.Fatalf("[%s] != [%s]", paths[0], path)
	}
}
