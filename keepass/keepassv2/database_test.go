package keepassv2_test

import (
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

	expected += r.Group.Name()
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}
}
