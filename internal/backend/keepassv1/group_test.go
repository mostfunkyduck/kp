package keepassv1_test

import (
	"strconv"
	"testing"

	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func TestGroupFunctions(t *testing.T) {
	ttlEntries := 50
	testName := "test name"
	db, err := keepass.New(&keepass.Options{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	group := db.Root().NewSubgroup()
	group.Name = testName

	for i := 0; i < ttlEntries; i++ {
		e, err := group.NewEntry()
		if err != nil {
			t.Fatalf(err.Error())
		}
		e.Title = "entry #" + strconv.Itoa(i)
		g := group.NewSubgroup()
		g.Name = "group #" + strconv.Itoa(i)
	}

	groupWrapper := v1.WrapGroup(group, &v1.Database{})
	// assuming stable ordering because the shell is premised on that for path traversal
	// (if the entries and groups change order, the user can't specify which one to change properly)
	for i, each := range groupWrapper.Groups() {
		name := "group #" + strconv.Itoa(i)
		if each.Name() != name {
			t.Errorf("%s != %s", each.Name(), name)
		}
		if each.Parent() == nil {
			t.Errorf("group %s had no parent!", each.Name())
		} else if each.Parent().Name() != testName {
			t.Errorf("parent name was incorrect for %s: %s", each.Name(), testName)
		}
	}

	for i, each := range groupWrapper.Entries() {
		name := "entry #" + strconv.Itoa(i)
		title := each.Title()
		if title != name {
			t.Errorf("%s != %s", title, name)
		}
	}
}
