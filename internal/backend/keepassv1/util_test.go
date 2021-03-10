package keepassv1_test

import (
	"testing"

	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func initDatabase() (t.Database, error) {
	// each 'key round' takes quite a while, make sure to use the minimum
	db, err := keepass.New(&keepass.Options{KeyRounds: 1})
	if err != nil {
		return nil, err
	}

	dbWrapper := v1.NewDatabase(db, "/dev/null")
	return dbWrapper, nil
}

func createTestResources(t *testing.T) runner.Resources {
	dbWrapper, err := initDatabase()
	if err != nil {
		t.Fatal(err)
	}
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
