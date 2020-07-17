package keepassv2

import (
	"encoding/base64"
	"fmt"
	k "github.com/mostfunkyduck/kp/keepass"
	gokeepasslib "github.com/tobischo/gokeepasslib/v3"
	"regexp"
)

type Group struct {
	group gokeepasslib.Group
	db    k.Database
}

func (g *Group) Raw() interface{} {
	return g.group
}

func newGroup(g gokeepasslib.Group, db k.Database) k.Group {
	return &Group{
		group: g,
		db:    db,
	}
}

func (g *Group) Groups() (rv []k.Group) {
	for _, each := range g.group.Groups {
		rv = append(rv, newGroup(each, g.db))
	}
	return
}

// findPathToGroup will attempt to find the path between 'source' and 'target', returning an
// ordered slice of groups leading up to *but not including* the target
// TODO this needs rigorous testing
func findPathToGroup(source k.Group, target k.Group) (rv []k.Group) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, group := range source.Groups() {
		uuidString, err := target.UUIDString()
		if err != nil {
			// TODO swallowing this error
			return []k.Group{}
		}
		groupUUIDString, err := group.UUIDString()
		if err != nil {
			return []k.Group{}
		}
		if groupUUIDString == uuidString {
			return append(rv, source)
		}
		if pathGroups := findPathToGroup(group, target); len(pathGroups) != 0 {
			return append([]k.Group{source}, rv...)
		}
	}
	return []k.Group{}
}

func (g *Group) Path() (rv string) {
	pathGroups := findPathToGroup(g.db.Root(), g)
	for _, each := range pathGroups {
		rv = rv + "/" + each.Name()
	}
	return rv + "/" + g.Name()
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, entry := range g.group.Entries {
		rv = append(rv, newEntry(entry, g.db))
	}
	return
}

func (g *Group) Parent() k.Group {
	if pathGroups := findPathToGroup(g.db.Root(), g); len(pathGroups) > 0 {
		return pathGroups[len(pathGroups)-1]
	}
	return nil
}

func (g *Group) SetParent(parent k.Group) error {
	rawParent := parent.Raw().(Group)
	rawParent.group.Groups = append(rawParent.group.Groups, g.group)
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
func (g *Group) NewSubgroup(name string) (k.Group, error) {
	newGroup := newGroup(gokeepasslib.NewGroup(), g.db)
	if err := newGroup.SetParent(g); err != nil {
		return &Group{}, fmt.Errorf("couldn't assign new group to parent '%s'; %s", g.Path(), err)
	}
	newGroup.SetName(name)
	return newGroup, nil
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
	subUUID, err := subgroup.UUIDString()
	if err != nil {
		return fmt.Errorf("could not read UUID on '%s': %s", subgroup.Path(), err)
	}

	for i, each := range g.group.Groups {
		eachWrapper := newGroup(each, g.db)
		eachUUID, err := eachWrapper.UUIDString()
		if err != nil {
			return fmt.Errorf("could not read UUID on '%s': %s", eachWrapper.Path(), err)
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

func (g *Group) NewEntry() (k.Entry, error) {
	entry := gokeepasslib.NewEntry()
	return &Entry{
		entry: entry,
	}, nil
}

func (g *Group) RemoveEntry(entry k.Entry) error {
	raw := g.group
	entryUUID, err := entry.UUIDString()
	if err != nil {
		return fmt.Errorf("cannot read UUID string on target entry '%s': %s", entry.Path(), err)
	}
	for i, each := range raw.Entries {
		eachWrapper := newEntry(each, g.db)
		eachUUID, err := eachWrapper.UUIDString()
		if err != nil {
			return fmt.Errorf("cannot read UUID string on individual entry '%s': %s", eachWrapper.Path(), err)
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
	for _, e := range g.group.Entries {
		// FIXME make e.Values part of the entry interface, this whole search shbang might be a util func
		for _, val := range e.Values {
			if term.FindString(val.Value.Content) != "" ||
				term.FindString(val.Key) != "" {
				// something in this entry matched, let's return it
				entryWrapper := newEntry(e, g.db)
				paths = append(paths, entryWrapper.Path())
			}
		}
	}
	return
}

// NOTE this is currently a copy of v1, might be something to make more general
func (g *Group) Search(term *regexp.Regexp) (paths []string) {

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
