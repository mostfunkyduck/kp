package keepassv1_test

import (
	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	runner "github.com/mostfunkyduck/kp/keepass/tests"
	"testing"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func createTestResources(t *testing.T) runner.Resources {
	// each 'key round' takes quite a while, make sure to use the minimum
	db, err := keepass.New(&keepass.Options{KeyRounds: 1})
	if err != nil {
		t.Fatalf(err.Error())
	}

	dbWrapper := v1.NewDatabase(db, "/dev/null")
	sg, err := dbWrapper.Root().NewSubgroup("asdf asdf asdf test")
	if err != nil {
		t.Fatalf(err.Error())
	}

	entry, err := sg.NewEntry("test test test")
	if err != nil {
		t.Fatalf(err.Error())
	}

	blankGroup, err := dbWrapper.Root().NewSubgroup("")
	if err != nil {
		t.Fatalf(err.Error())
	}
	blankEntry, err := blankGroup.NewEntry("")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// NOTE this library doesn't support blank objects, so the blanks are just separate groups
	return runner.Resources{
		Db:         dbWrapper,
		Entry:      entry,
		Group:      sg,
		BlankEntry: blankEntry,
		BlankGroup: blankGroup,
	}
}
