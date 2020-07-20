package keepassv2_test

import (
	"regexp"
	"testing"

	k "github.com/mostfunkyduck/kp/keepass"
)

func findGroupInGroup(parent k.Group, child k.Group) bool {
	for _, group := range parent.Groups() {
		if group.Name() == child.Name() {
			return true
		}
	}
	return false
}

func TestGroupFunctions(t *testing.T) {
	r := createTestResources(t)
	root := r.Db.Root()
	name := "TestGroupFunctions"
	sg, err := root.NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	found := findGroupInGroup(root, sg)
	if !found {
		t.Fatalf("added a subgroup, but couldn't find it afterwards in root's groups")
	}

	originalGroupCount := len(root.Groups())
	_, err = root.NewSubgroup(name)
	if err == nil {
		t.Fatalf("added duplicate subgroup to root")
	}

	newGroupCount := len(root.Groups())
	if originalGroupCount != newGroupCount {
		t.Fatalf("%d != %d", originalGroupCount, newGroupCount)
	}

	if err := root.RemoveSubgroup(sg); err != nil {
		t.Fatalf(err.Error())
	}

	if findGroupInGroup(root, sg) {
		t.Fatalf("found group '%s' even after it was removed from too", sg.Name())
	}

	if err := root.RemoveSubgroup(sg); err == nil {
		t.Fatalf("was able to remove subgroup twice")
	}
}

func TestParentFunctions(t *testing.T) {
	r := createTestResources(t)
	if err := r.Db.Root().SetParent(r.Group); err == nil {
		t.Fatalf("was able to set root's parent")
	}

	p := r.Db.Root().Parent()
	if p != nil {
		t.Fatalf("%v", p)
	}
}

func TestRootGroupIsRoot(t *testing.T) {
	r := createTestResources(t)
	if !r.Db.Root().IsRoot() {
		t.Fatalf("IsRoot is broken")
	}
}

func TestEntryFunctions(t *testing.T) {
	r := createTestResources(t)
	e, err := r.Group.NewEntry("test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if err := r.Db.Root().AddEntry(e); err == nil {
		t.Fatalf("shouldn't be able to add an entry to root")
	}

	if err := r.Db.Root().RemoveEntry(e); err == nil {
		t.Fatalf("shouldn't be able to remove entry from root")
	}

	entriesLen := len(r.Db.Root().Entries())
	if entriesLen != 0 {
		t.Fatalf("found %d entries in root group", entriesLen)
	}
}

func TestSearch(t *testing.T) {
	name := "askldfjhasl;kcvjs;lkjnsfasdlfkjas;ldvkjsdl;fgkja;skdlfgjnw;oeihaw;oifhjas;kldfjhasf"
	r := createTestResources(t)
	sg, err := r.Db.Root().NewSubgroup(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// search for group using partial search
	paths := r.Db.Root().Search(regexp.MustCompile(name[0 : len(name)/2]))
	if len(paths) != 1 {
		t.Fatalf("too many paths returned from group search: %v", paths)
	}

	path, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if paths[0] != path+"/" {
		t.Fatalf("[%s] != [%s]", paths[0], path)
	}

	name = r.Entry.Title()
	paths = r.Db.Root().Search(regexp.MustCompile(name[0 : len(name)/2]))
	if len(paths) != 1 {
		t.Fatalf("wrong count of paths returned from entry search: %v", paths)
	}

	path, err = r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if paths[0] != path {
		t.Fatalf("[%s] != [%s]", paths[0], path)
	}
}
