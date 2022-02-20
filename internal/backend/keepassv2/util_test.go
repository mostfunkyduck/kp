package keepassv2_test

import (
	"os"
	"testing"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	main "github.com/mostfunkyduck/kp/internal/backend/keepassv2"
	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
	"github.com/mostfunkyduck/kp/internal/backend/types"
	g "github.com/tobischo/gokeepasslib/v3"
)

func createTestResources(t *testing.T) runner.Resources {
	name := "test yo"
	groupName := "group"
	db := &main.Database{}
	tmpfile, err := os.CreateTemp("", "kp_unit_tests")
	if err != nil {
		t.Fatal(err)
	}
	fileName := tmpfile.Name()
	defer os.Remove(fileName)

	// remove the file before we init the DB so that it will create the file from scratch
	// we're really only creating the tempfile as a shortcut to generate an appropriate path
	tmpfile.Close()
	os.Remove(fileName)

	opts := types.Options{
		DBPath: fileName,
	}
	if err := db.Init(opts); err != nil {
		t.Fatal(err)
	}
	newgrp := g.NewGroup()
	group := main.WrapGroup(&newgrp, db)
	group.SetName(groupName)
	if err := db.Root().AddSubgroup(group); err != nil {
		t.Fatal(err)
	}
	newEnt := g.NewEntry()
	entry := main.WrapEntry(&newEnt, db)
	if !entry.Set(c.NewValue(
		[]byte(name),
		"Title",
		false, false, false,
		types.STRING,
	)) {
		t.Fatalf("could not set title")
	}
	if err := entry.SetParent(group); err != nil {
		t.Fatalf(err.Error())
	}

	rawEnt := g.NewEntry()
	rawGrp := g.NewGroup()

	return runner.Resources{
		Db:         db,
		Group:      group,
		Entry:      entry,
		BlankEntry: main.WrapEntry(&rawEnt, db),
		BlankGroup: main.WrapGroup(&rawGrp, db),
	}
}
