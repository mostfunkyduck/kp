package keepassv2_test

import (
	"testing"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	main "github.com/mostfunkyduck/kp/internal/backend/keepassv2"
	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
	types "github.com/mostfunkyduck/kp/internal/backend/types"
	g "github.com/tobischo/gokeepasslib/v3"
)

func createTestResources(t *testing.T) runner.Resources {
	name := "test yo"
	groupName := "group"
	db := main.NewDatabase(g.NewDatabase(), "/dev/null", types.Options{})
	newgrp := g.NewGroup()
	group := main.WrapGroup(&newgrp, db)
	group.SetName(groupName)
	if err := db.Root().AddSubgroup(group); err != nil {
		t.Fatalf(err.Error())
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
