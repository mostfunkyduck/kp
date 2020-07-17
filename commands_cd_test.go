package main_test

import (
	"testing"

	main "github.com/mostfunkyduck/kp"
)

func TestCdToGroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{
		r.Group.Name(),
	}

	r.Db.SetCurrentLocation(r.Db.Root())
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	if currentLocation.Path() != r.Group.Path() {
		t.Fatalf("new location was not the one specified: %s != %s", currentLocation.Path(), r.Group.Path())
	}
}

func TestCdToRoot(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}

	r.Db.SetCurrentLocation(r.Group)
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	if currentLocation.Path() != r.Db.Root().Path() {
		t.Fatalf("new location was not the one specified: %s != %s", currentLocation.Path(), r.Db.Root().Path())
	}
}
