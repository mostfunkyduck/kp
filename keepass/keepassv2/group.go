package keepassv2

import (
	"encoding/base64"
	"fmt"

	k "github.com/mostfunkyduck/kp/keepass"
	c "github.com/mostfunkyduck/kp/keepass/common"
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
func WrapGroup(g *gokeepasslib.Group, db k.Database) k.Group {
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

func (g *Group) Groups() (rv []k.Group) {
	for _, each := range g.group.Groups {
		_each := each
		rv = append(rv, WrapGroup(&_each, g.DB()))
	}
	return
}

// findPathToGroup will attempt to find the path between 'source' and 'target', returning an
// ordered slice of groups leading up to *but not including* the target
// TODO this needs rigorous testing
func findPathToGroup(source k.Group, target k.Group) (rv []k.Group, err error) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, group := range source.Groups() {
		same, err := CompareUUIDs(group, target)
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not compare UUIDS: %s", err)
		}

		if same {
			return []k.Group{source}, nil
		}

		pathGroups, err := findPathToGroup(group, target)
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not find path from group '%s' to group '%s': %s", group.Name(), target.Name(), err)
		}

		if len(pathGroups) != 0 {
			return append([]k.Group{source}, pathGroups...), nil
		}
	}
	return []k.Group{}, nil
}

func (g *Group) Path() (rv string, err error) {
	pathGroups, err := findPathToGroup(g.DB().Root(), g)
	if err != nil {
		return rv, fmt.Errorf("could not find path to group '%s'", g.Name())
	}
	for _, each := range pathGroups {
		rv = rv + each.Name() + "/"
	}
	return rv + g.Name() + "/", nil
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, entry := range g.group.Entries {
		_entry := entry
		rv = append(rv, WrapEntry(&_entry, g.DB()))
	}
	return
}

func (g *Group) Parent() k.Group {
	pathGroups, err := findPathToGroup(g.DB().Root(), g)
	if err != nil {
		return nil
	}
	if len(pathGroups) > 0 {
		return pathGroups[len(pathGroups)-1]
	}
	return nil
}

func (g *Group) SetParent(parent k.Group) error {
	// Since there's no child->parent relationship in this library, we need
	// to rely on the parent->child connection to get this to work
	return parent.AddSubgroup(g)
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
func (g *Group) NewSubgroup(name string) (k.Group, error) {
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

func (g *Group) AddSubgroup(subgroup k.Group) error {
	for _, each := range g.Groups() {
		if each.Name() == subgroup.Name() {
			return fmt.Errorf("group named '%s' already exists", each.Name())
		}
	}

	g.group.Groups = append(g.group.Groups, *subgroup.Raw().(*gokeepasslib.Group))
	subgroup.(*Group).updateWrapper(&g.group.Groups[len(g.group.Groups)-1])
	return nil
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
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

func (g *Group) AddEntry(e k.Entry) error {
	for _, each := range g.Entries() {
		if each.Title() == e.Title() {
			return fmt.Errorf("entry named '%s' already exists", each.Title())
		}
	}
	g.group.Entries = append(g.group.Entries, *e.Raw().(*gokeepasslib.Entry))
	// TODO update entry wrapper
	return nil
}
func (g *Group) NewEntry(name string) (k.Entry, error) {
	entry := gokeepasslib.NewEntry()
	entryWrapper := WrapEntry(&entry, g.DB())
	entryWrapper.SetTitle(name)
	entryWrapper.Set(k.Value{
		Name:  "UserName",
		Value: []byte(""),
	})
	entryWrapper.SetPassword("")

	if err := entryWrapper.SetParent(g); err != nil {
		return nil, fmt.Errorf("could not add entry to group: %s", err)
	}
	return entryWrapper, nil
}

func (g *Group) RemoveEntry(entry k.Entry) error {
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
