package keepassv1_test

import (
	"testing"

	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	"zombiezen.com/go/sandpass/pkg/keepass"
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
	expectedName := "asdf"
	db := &v1.Database{}
	kdbGroup := v1.WrapGroup(&keepass.Group{Name: expectedName}, db)
	db.SetCurrentLocation(kdbGroup)
	l := db.CurrentLocation()
	if l == nil {
		t.Fatalf("could not retrieve current location")
	}
	name := l.Name()
	if name != expectedName {
		t.Fatalf("%s != %s", name, expectedName)
	}

}
