package keepassv1_test

import (
	"testing"

	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
)

func TestSavePath(t *testing.T) {
	sp := "adsfasdfjkalskdfj"
	db := &v1.Database{}

	db.SetSavePath(sp)
	dbSp := db.SavePath()
	if sp != dbSp {
		t.Errorf("%s != %s", sp, dbSp)
	}
}

func TestCurrentLocation(t *testing.T) {
	r := createTestResources(t)
	expectedName := "asdf"
	newGroup, err := r.Group.NewSubgroup(expectedName)
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Db.SetCurrentLocation(newGroup)
	l := r.Db.CurrentLocation()
	if l == nil {
		t.Fatalf("could not retrieve current location")
	}
	name := l.Name()
	if name != expectedName {
		t.Fatalf("%s != %s", name, expectedName)
	}

}

func TestBinaries(t *testing.T) {
	r := createTestResources(t)
	b, _ := r.Db.Binary(10000, "blork blork")
	if b.Present {
		t.Fatal("got a binary from the v1 DB, this is not supported by v1, so it's a mystery to me")
	}
}
