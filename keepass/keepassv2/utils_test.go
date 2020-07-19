package keepassv2_test

import (
	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/keepassv2"
	g "github.com/tobischo/gokeepasslib/v3"
	"testing"
)

type Resources struct {
	Db    k.Database
	Group k.Group
	Entry k.Entry
}

func createTestResources(t *testing.T) Resources {
	name := "test yo"
	groupName := "group"
	db := main.NewDatabase(g.NewDatabase(), "/dev/null", k.Options{})
	newgrp := g.NewGroup()
	group := main.WrapGroup(&newgrp, db)
	group.SetName(groupName)
	if err := db.Root().AddSubgroup(group); err != nil {
		t.Fatalf(err.Error())
	}
	newEnt := g.NewEntry()
	entry := main.WrapEntry(&newEnt, db)
	if err := entry.SetParent(group); err != nil {
		t.Fatalf(err.Error())
	}
	if !entry.Set(k.Value{Name: "Title", Value: name}) {
		t.Fatalf("could not set title")
	}
	return Resources{
		Db:    db,
		Group: group,
		Entry: entry,
	}
}
