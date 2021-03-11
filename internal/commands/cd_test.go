package commands_test

import (
	"testing"

	main "github.com/mostfunkyduck/kp/internal/commands"
)

func TestCdToGroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{
		r.Group.Name(),
	}

	r.Db.SetCurrentLocation(r.Db.Root())
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	clPath, err := currentLocation.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	rGrpPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if clPath != rGrpPath {
		t.Fatalf("new location was not the one specified: %s != %s", clPath, rGrpPath)
	}
}

func TestCdToRoot(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}

	r.Db.SetCurrentLocation(r.Group)
	main.Cd(r.Shell)(r.Context)

	currentLocation := r.Db.CurrentLocation()
	clPath, err := currentLocation.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	rDbPath, err := r.Db.Root().Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if clPath != rDbPath {
		t.Fatalf("new location was not the one specified: %s != %s", clPath, rDbPath)
	}
}
