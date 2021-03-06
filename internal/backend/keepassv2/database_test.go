package keepassv2_test

import (
	"regexp"
	"testing"
)

func TestDbPath(t *testing.T) {
	r := createTestResources(t)
	path, err := r.Db.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "/"
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}

	r.Db.SetCurrentLocation(r.Group)

	path, err = r.Db.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected += r.Group.Name() + "/"
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}
}

func TestDBSearch(t *testing.T) {
	r := createTestResources(t)
	paths, err := r.Db.Search(regexp.MustCompile(r.Group.Name()))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("%v", paths)
	}

	paths, err = r.Db.Search(regexp.MustCompile(r.Entry.Title()))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("%v", paths)
	}
}
