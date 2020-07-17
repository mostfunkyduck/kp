package main_test

import (
	"regexp"
	"testing"
)

func TestSearchFullPath(t *testing.T) {
	r := createTestResources(t)
	term, err := regexp.Compile(r.Entry.Get("title").Value.(string))
	if err != nil {
		t.Fatalf(err.Error())
	}
	paths := r.Group.Search(term)
	if len(paths) != 1 {
		t.Fatalf("%d != %d", len(paths), 1)
	}

	if paths[0] != r.Entry.Path() {
		t.Fatalf("[%s] != [%s]", paths[0], r.Entry.Path())
	}
}
