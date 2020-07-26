package keepassv1

import (
	"fmt"
	k "github.com/mostfunkyduck/kp/keepass"
	"regexp"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Group struct {
	db k.Database
	group *keepass.Group
}

func WrapGroup(group *keepass.Group, db k.Database) k.Group {
	return &Group{
		db: db,
		group: group,
	}
}

func (g *Group) AddSubgroup(subgroup k.Group) error {
	if err := subgroup.SetParent(g); err != nil {
		return fmt.Errorf("could not set subgroup parent: %s", err)
	}
	return nil
}

func (g *Group) AddEntry(e k.Entry) error {
	if err := e.SetParent(g); err != nil {
		return fmt.Errorf("could not add entry: %s", err)
	}
	return nil
}

// FIXME the keepass library has a bug where you can't get the parent
// unless the entry is a pointer to the one in the db (it's comparing pointer values)
// this can/should/will be fixed in my fork
func (g *Group) searchEntries(term *regexp.Regexp) (paths []string) {
	for _, e := range g.Entries() {
		if term.FindString(e.Get("title").Value.(string)) != "" ||
			term.FindString(e.Get("notes").Value.(string)) != "" ||
			term.FindString(e.Get("attachment").Name) != "" ||
			term.FindString(e.Get("username").Value.(string)) != "" {
			path, err := e.Path()
			if err != nil {
				path = fmt.Sprintf("<error: %s>", err)
			}
			paths = append(paths, path)
		}
	}
	return
}

func (g *Group) Search(term *regexp.Regexp) (paths []string) {

	paths = append(paths, g.searchEntries(term)...)

	for _, g := range g.Groups() {
		paths = append(paths, g.Search(term)...)
	}
	return paths
}

func (g *Group) Name() string {
	return g.group.Name
}

func (g *Group) SetName(name string) {
	g.group.Name = name
}

func (g *Group) Parent() k.Group {
	return WrapGroup(g.group.Parent(), g.db)
}

func (g *Group) SetParent(parent k.Group) error {
	if err := g.group.SetParent(parent.Raw().(*keepass.Group)); err != nil {
		return fmt.Errorf("could not change group parent: %s", err)
	}
	return nil
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, each := range g.group.Entries() {
		rv = append(rv, &Entry{entry: each})
	}
	return rv
}

func (g *Group) Groups() (rv []k.Group) {
	for _, each := range g.group.Groups() {
		rv = append(rv, WrapGroup(each, g.db))
	}
	return rv
}

func (g *Group) IsRoot() bool {
	return g.Parent() == nil
}

func (g *Group) NewSubgroup(name string) (k.Group, error) {
	newGroup := g.group.NewSubgroup()
	newGroup.Name = name
	return WrapGroup(newGroup, g.db), nil
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
	return g.group.RemoveSubgroup(subgroup.Raw().(*keepass.Group))
}

func (g *Group) Path() (fullPath string, err error) {
	// FIXME this shouldn't need access to the bare group
	group := g.group
	for ; group != nil; group = group.Parent() {
		if group.IsRoot() {
			fullPath = "/" + fullPath
			break
		}
		fullPath = group.Name + "/" + fullPath
	}
	return fullPath, nil
}

func (g *Group) Raw() interface{} {
	return g.group
}

func (g *Group) NewEntry(name string) (k.Entry, error) {
	entry, err := g.group.NewEntry()
	if err != nil {
		return nil, err
	}
	entry.Title = name
	return &Entry{
		entry: entry,
	}, nil
}

func (g *Group) RemoveEntry(e k.Entry) error {
	return g.group.RemoveEntry(e.Raw().(*keepass.Entry))
}

func (g *Group) UUIDString() (string, error) {
	return string(g.group.ID), nil
}

func (g *Group) DB() k.Database {
	return g.db
}
