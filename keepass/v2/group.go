package keepassv2

import (
	"regexp"
	k "github.com/mostfunkyduck/kp/keepass"
	gokeepasslib "github.com/tobischo/gokeepasslib/v3"
)

type Group struct {
	group gokeepasslib.Group
	db	k.Database
}
func (g *Group) Raw() interface{} {
	return g.group
}

func newGroup(g gokeepasslib.Group, db k.Database) k.Group {
	return Group{
		group: g,
		db: db,
	}
}

func (g *Group) Groups() (rv []k.Group) {
	for _, each := range g.group.Groups {
		rv = append(rv, newGroup(each, g.db))
	}
}

// findPathToGroup will attempt to find the path between 'source' and 'target', returning an
// ordered slice of groups leading up to the target
// TODO this needs rigorous testing
func findPathToGroup(source k.Group, target k.Group) (rv []Group){
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, group := range source.Groups() {
		if group.UUIDString() == target.UUIDString() {
			return append(rv, group)
		}
		if pathGroups := findPathToGroup(group, target); len(pathGroups) != 0 {
			return append([]Group{group}, rv...)
		}
	}
	return []Group{}
}
func (g *Group) Path() string {
	findPathToGroup(g.db.Root(), g)
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, entry := range g.Raw().Entries {
		rv = append(rv, newEntry(entry))
	}
	return
}

func (g *Group) Parent() (k.Group) {
	if pathGroups := findPathToGroup(g.db, g); len(pathGroups) > 0 {
		// the parent will be second from the end
		return pathGroups[len(pathGroups)-2]
	}
	return nil
}

func (g *Group) SetParent(parent k.Group) error {
	parent.Raw().Groups = append(parent.Raw().Groups, g.Raw())
	return nil
}

func (g *Group) Name() string {
	return g.Raw().Name
}

func (g *Group) SetName(name string) {
	g.Raw().Name = name
}

func (g *Group) IsRoot() bool {
	// there's a separate struct for root, this one is always used for subgroups
	return false
}
	// Creates a new subgroup with a given name under this group
func (g *Group) NewSubgroup(name string) k.Group {
	newGroup := gokeepasslib.NewGroup([]gokeepasslib.Options{})
	newGroup.SetParent(g)
	return newGroup
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
	for i, each := range g.Raw().Groups {
		if each.UUIDString() == subgroup.UUIDString() {
			// remove it
			raw := g.Raw()
			groupLen := len(raw.Groups)
			raw.Groups = append(raw.Groups[0:i], raw.Groups[i+1:groupLen])
			return nil
		}
	}
	return fmt.Errorf("could not find group with UUID '%s'", subgroup.UUIDString())
}

func (g *Group) NewEntry() (k.Entry, error) {
	entry := g.NewEntry()
	return Entry{
		entry: entry,
	}, nil
}

func (g *Group) RemoveEntry(entry k.Entry) error {
	raw := g.Raw()
	for i, each := range raw.Entries {
		if each.UUIDString() == entry.UUIDString() {
			entriesLen := len(raw.Entries)
			raw.Entries = append(raw.Entries[0:i], raw.Entries[i+1:entriesLen])
			return nil
		}
	}
	return fmt.Errorf("could not find entry with UUID '%s'", entry.UUIDString())
}

func (g *Group) searchEntries(term *regexp.Regexp) (titles []string) {
	for _, e := range g.Raw().Entries {
		// FIXME make e.Values part of the entry interface, this whole search shbang might be a util func
		for _, val := range e.Values {
			if term.FindString(val.V.Content) != "" ||
				 term.FindString(val.Key) != "" {
					titles
				 }
			 }
		}
	return titles
}

// NOTE this is currently a copy of v1, might be something to make more general
func (g *Group) Search(term *regexp.Regexp) (paths []string) {

	for _, title := range g.searchEntries(term) {
		paths = append(paths, "./"+title)
	}
	for _, g := range g.Groups() {
		paths = append(paths, g.Search(term)...)
	}
	return paths
}
