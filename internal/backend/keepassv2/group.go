package keepassv2

import (
	"encoding/base64"
	"fmt"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	gokeepasslib "github.com/tobischo/gokeepasslib/v3"
)

type Group struct {
	c.Group
	group *gokeepasslib.Group
}

func (g *Group) Raw() interface{} {
	return g.group
}

// WrapGroup wraps a bare gokeepasslib.Group and a database in a Group wrapper
func WrapGroup(g *gokeepasslib.Group, db t.Database) t.Group {
	if g == nil {
		return nil
	}
	gr := &Group{
		group: g,
	}
	gr.SetDB(db)
	gr.SetDriver(gr)
	return gr
}

func (g *Group) Groups() (rv []t.Group) {
	for i := range g.group.Groups {
		rv = append(rv, WrapGroup(&g.group.Groups[i], g.DB()))
	}
	return
}

func (g *Group) Entries() (rv []t.Entry) {
	for i := range g.group.Entries {
		rv = append(rv, WrapEntry(&g.group.Entries[i], g.DB()))
	}
	return
}

func (g *Group) Parent() t.Group {
	pathGroups, err := c.FindPathToGroup(g.DB().Root(), g)
	if err != nil {
		return nil
	}
	if len(pathGroups) > 0 {
		return pathGroups[len(pathGroups)-1]
	}
	return nil
}

func (g *Group) SetParent(parent t.Group) error {
	oldParent := g.Parent()

	// If the group is being renamed, the parents will be the same
	if oldParent != nil {
		sameParent, err := c.CompareUUIDs(oldParent, parent)
		if err != nil {
			return fmt.Errorf("error comparing new parent UUID to old parent UUID: %s", err)
		}
		if sameParent {
			return nil
		}
	}

	// Since there's no child->parent relationship in this library, we need
	// to rely on the parent->child connection to get this to work
	if err := parent.AddSubgroup(g); err != nil {
		return err
	}

	// Since "parent" is defined as "being in a group's subgroup", the group may now have two of them,
	// we need to make sure it's only in one
	if oldParent != nil {
		if err := oldParent.RemoveSubgroup(g); err != nil {
			// the group doesn't exist in the parent anymore or the UUIDs got corrupted
			// stop at this point since something got corrupted
			return fmt.Errorf("error removing group from existing parent, possible data corruption has occurred: %s", err)
		}
	}
	return nil
}

func (g *Group) Name() string {
	return g.group.Name
}

func (g *Group) SetName(name string) {
	g.group.Name = name
}

func (g *Group) IsRoot() bool {
	// there's a separate struct for root, this one is always used for subgroups
	return false
}

// Creates a new subgroup with a given name under this group
func (g *Group) NewSubgroup(name string) (t.Group, error) {
	newGroup := gokeepasslib.NewGroup()
	newGroupWrapper := WrapGroup(&newGroup, g.DB())
	newGroupWrapper.SetName(name)
	if err := newGroupWrapper.SetParent(g); err != nil {
		return &Group{}, fmt.Errorf("couldn't assign new group to parent '%s'; %s", g.Name(), err)
	}
	return newGroupWrapper, nil
}

func (g *Group) updateWrapper(group *gokeepasslib.Group) {
	g.group = group
}

func (g *Group) AddSubgroup(subgroup t.Group) error {
	for _, each := range g.Groups() {
		if each.Name() == subgroup.Name() {
			return fmt.Errorf("group named '%s' already exists", each.Name())
		}
	}

	g.group.Groups = append(g.group.Groups, *subgroup.Raw().(*gokeepasslib.Group))
	subgroup.(*Group).updateWrapper(&g.group.Groups[len(g.group.Groups)-1])
	return nil
}

// RemoveSubgroup will remove a group from a parent group
// If this function returns an error, that means that either the UUIDs on the parent or child
// were corrupted or the group didn't actually exist in the parent
func (g *Group) RemoveSubgroup(subgroup t.Group) error {
	subUUID, err := subgroup.UUIDString()
	if err != nil {
		return fmt.Errorf("could not read UUID on subgroup '%s': %s", subgroup.Name(), err)
	}

	for i, each := range g.group.Groups {
		eachWrapper := WrapGroup(&each, g.DB())
		eachUUID, err := eachWrapper.UUIDString()
		if err != nil {
			return fmt.Errorf("could not read UUID on '%s': %s", eachWrapper.Name(), err)
		}

		if eachUUID == subUUID {
			// remove it
			raw := g.group
			groupLen := len(raw.Groups)
			raw.Groups = append(raw.Groups[0:i], raw.Groups[i+1:groupLen]...)
			return nil
		}
	}
	return fmt.Errorf("could not find group with UUID '%s'", subUUID)
}

func (g *Group) AddEntry(e t.Entry) error {
	for _, each := range g.Entries() {
		if each.Title() == e.Title() {
			return fmt.Errorf("entry named '%s' already exists", each.Title())
		}
	}
	g.group.Entries = append(g.group.Entries, *e.Raw().(*gokeepasslib.Entry))
	// TODO update entry wrapper
	return nil
}
func (g *Group) NewEntry(name string) (t.Entry, error) {
	entry := gokeepasslib.NewEntry()
	entryWrapper := WrapEntry(&entry, g.DB())
	// the order in which these values are added determines how they are output in the terminal
	// both for prompts and output
	entryWrapper.SetTitle(name)
	entryWrapper.Set(c.NewValue(
		[]byte(""),
		"URL",
		true, false, false,
		t.STRING,
	))
	entryWrapper.Set(c.NewValue(
		// This needs to be formatted this way to tie in to how keepass2 looks for usernames
		[]byte(""),
		"UserName",
		true, false, false,
		t.STRING,
	))
	entryWrapper.SetPassword("")
	entryWrapper.Set(c.NewValue(
		[]byte(""),
		"Notes",
		true, false, false,
		t.LONGSTRING,
	))
	if err := entryWrapper.SetParent(g); err != nil {
		return nil, fmt.Errorf("could not add entry to group: %s", err)
	}
	return entryWrapper, nil
}

func (g *Group) RemoveEntry(entry t.Entry) error {
	raw := g.group
	entryUUID, err := entry.UUIDString()
	if err != nil {
		return fmt.Errorf("cannot read UUID string on target entry '%s': %s", entry.Title(), err)
	}
	for i, each := range raw.Entries {
		eachWrapper := WrapEntry(&each, g.DB())
		eachUUID, err := eachWrapper.UUIDString()
		if err != nil {
			return fmt.Errorf("cannot read UUID string on individual entry '%s': %s", eachWrapper.Title(), err)
		}
		if eachUUID == entryUUID {
			entriesLen := len(raw.Entries)
			raw.Entries = append(raw.Entries[0:i], raw.Entries[i+1:entriesLen]...)
			return nil
		}
	}
	return fmt.Errorf("could not find entry with UUID '%s'", entryUUID)
}

func (g *Group) UUIDString() (string, error) {
	encodedUUID, err := g.group.UUID.MarshalText()
	if err != nil {
		return "", fmt.Errorf("could not encode UUID: %s", err)
	}
	str, err := base64.StdEncoding.DecodeString(string(encodedUUID))
	if err != nil {
		return "", fmt.Errorf("could not decode b64: %s", err)
	}
	return string(str), nil
}
