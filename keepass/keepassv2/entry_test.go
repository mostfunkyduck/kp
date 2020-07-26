package keepassv2_test

import (
	"testing"

	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/keepassv2"
	g "github.com/tobischo/gokeepasslib/v3"
	runner "github.com/mostfunkyduck/kp/keepass/tests"
)

func TestNoParent(t *testing.T) {
	r := runner.Resources{}
	r.Db = main.NewDatabase(g.NewDatabase(), "/dev/null", k.Options{})
	newEnt := g.NewEntry()
	r.Entry = main.WrapEntry(&newEnt, r.Db)

	runner.RunTestNoParent(t, r)
}

func TestRegularPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestRegularPath(t, r)
}

func TestEntryGetSet (t *testing.T) {
	r := createTestResources(t)
	runner.RunTestGetSet(t, r)
}

func TestEntryTimeFuncs (t *testing.T) {
	r := createTestResources(t)
	runner.RunTestEntryTimeFuncs(t, r)
}

func TestEntryPasswordTitleFuncs (t *testing.T) {
	r := createTestResources(t)
	runner.RunTestEntryPasswordTitleFuncs(t, r)
}

func TestOutput (t *testing.T) {
	r := createTestResources(t)
	runner.RunTestOutput(t, r)
}
