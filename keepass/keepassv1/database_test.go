package keepassv1_test

import (
	"testing"

	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
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
