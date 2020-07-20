package keepassv2

import (
	"encoding/base64"
	"fmt"
	k "github.com/mostfunkyduck/kp/keepass"
	gokeepasslib "github.com/tobischo/gokeepasslib/v3"
	"regexp"
)

type Group struct {
	group *gokeepasslib.Group
	db    k.Database
}

func (g *Group) Raw() interface{} {
	return g.group
}

// WrapGroup wraps a bare gokeepasslib.Group and a database in a Group wrapper
func WrapGroup(g *gokeepasslib.Group, db k.Database) k.Group {
	return &Group{
		group: g,
		db:    db,
	}
}

func (g *Group) Groups() (rv []k.Group) {
	for _, each := range g.group.Groups {
		_each := each
		rv = append(rv, WrapGroup(&_each, g.db))
	}
	return
}

// findPathToGroup will attempt to find the path between 'source' and 'target', returning an
// ordered slice of groups leading up to *but not including* the target
// TODO this needs rigorous testing
func findPathToGroup(source k.Group, target k.Group) (rv []k.Group, err error) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, group := range source.Groups() {
		uuidString, err := target.UUIDString()
		if err != nil {
			// TODO swallowing this error
			return []k.Group{}, fmt.Errorf("could not parse UUID string in group '%s'", target.Name())
		}
		groupUUIDString, err := group.UUIDString()
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not parse UUID string in group '%s'", group.Name())
		}
		if groupUUIDString == uuidString {
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
	pathGroups, err := findPathToGroup(g.db.Root(), g)
	if err != nil {
		return rv, fmt.Errorf("could not find path to group '%s'", g.Name())
	}
	for _, each := range pathGroups {
		rv = rv + each.Name() + "/"
	}
	return rv + g.Name(), nil
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, entry := range g.group.Entries {
		_entry := entry
		rv = append(rv, WrapEntry(&_entry, g.db))
	}
	return
}

func (g *Group) Parent() k.Group {
	pathGroups, err := findPathToGroup(g.db.Root(), g)
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
	newGroupWrapper := WrapGroup(&newGroup, g.db)
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

	for _, each := range g.Entries() {
		if each.Title() == subgroup.Name() {
			return fmt.Errorf("entry named '%s' already exists", each.Title())
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
		eachWrapper := WrapGroup(&each, g.db)
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
	entryWrapper := WrapEntry(&entry, g.db)
	entryWrapper.Set(k.Value{Name: "Title", Value: name})
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
		eachWrapper := WrapEntry(&each, g.db)
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

func (g *Group) searchEntries(term *regexp.Regexp) (paths []string) {
	for _, e := range g.Entries() {
		// FIXME make e.Values part of the entry interface, this whole search shbang might be a util func
		for _, val := range e.Values() {
			content := val.Value.(string)
			if term.FindString(content) != "" ||
				term.FindString(val.Name) != "" {
				// something in this entry matched, let's return it
				path, _ := e.Path()
				paths = append(paths, path)
			}
		}
	}
	return
}

// NOTE this is currently a copy of v1, might be something to make more general
func (g *Group) Search(term *regexp.Regexp) (paths []string) {
	if term.FindString(g.Name()) != "" {
		path, err := g.Path()
		if err == nil {
			// append slash so it's clear that it's a group, not an entry
			paths = append(paths, path + "/")
		}
	}
	paths = append(paths, g.searchEntries(term)...)
	for _, g := range g.Groups() {
		paths = append(paths, g.Search(term)...)
	}
	return paths
}

// FIXME make this a library function, should not require effort
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
