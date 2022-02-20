package keepassv1_test

import (
	"fmt"
	"os"
	"testing"

	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func initDatabase() (t.Database, error) {
	dbWrapper := &v1.Database{}
	// yes, unit tests should avoid the file system.  baby steps.
	tmpfile, err := os.CreateTemp("", "kp_unit_tests")
	if err != nil {
		return dbWrapper, fmt.Errorf("could not create temp file for DB: %s", tmpfile.Name())
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())
	defer os.Remove(tmpfile.Name())

	dbOptions := t.Options{
		DBPath: tmpfile.Name(),
		// each 'key round' takes quite a while, make sure to use the minimum
		KeyRounds: 1,
	}
	if err := dbWrapper.Init(dbOptions); err != nil {
		return dbWrapper, fmt.Errorf("could not init db with provided options: %s", err)
	}
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
