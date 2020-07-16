package main_test

import (
	main "github.com/mostfunkyduck/kp"
	"testing"
)
func TestCdToGroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{
		r.Group.Name(),
	}

	r.Db.SetCurrentLocation(r.Db.Root())
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	if currentLocation.Pwd() != r.Group.Pwd() {
		t.Fatalf("new location was not the one specified: %s != %s", currentLocation.Pwd(), r.Group.Pwd())
	}
}

func TestCdToRoot(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}

	r.Db.SetCurrentLocation(r.Group)
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	if currentLocation.Pwd() != r.Db.Root().Pwd() {
		t.Fatalf("new location was not the one specified: %s != %s", currentLocation.Pwd(), r.Db.Root().Pwd())
	}
}

