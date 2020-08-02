package keepassv1

import (
	"fmt"
	"regexp"

	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Group struct {
	db    k.Database
	group *keepass.Group
}

func WrapGroup(group *keepass.Group, db k.Database) k.Group {
	if group == nil {
		return nil
	}
	g := &Group{
		db:    db,
		group: group,
	}

	return g
}

func (g *Group) AddSubgroup(subgroup k.Group) error {
	for _, group := range g.Groups() {
		if group.Name() == subgroup.Name() {
			return fmt.Errorf("group named '%s' already exists at this location", group.Name())
		}
	}
	if err := subgroup.SetParent(g); err != nil {
		return fmt.Errorf("could not set subgroup parent: %s", err)
	}
	return nil
}

func (g *Group) AddEntry(e k.Entry) error {
	for _, each := range g.Entries() {
		if each.Title() == e.Title() {
			return fmt.Errorf("entry named '%s' already exists at this location", e.Title())
		}
	}
	if err := e.SetParent(g); err != nil {
		return fmt.Errorf("could not add entry: %s", err)
	}
	return nil
}

func (g *Group) Search(term *regexp.Regexp) (paths []string) {
	if term.FindString(g.Name()) != "" {
		path, err := g.Path()
		if err == nil {
			// append slash so it's clear that it's a group, not an entry
			paths = append(paths, path+"/")
		}
	}

	for _, e := range g.Entries() {
		paths = append(paths, e.Search(term)...)
	}

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
		rv = append(rv, WrapEntry(each, g.DB()))
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
	for _, group := range g.Groups() {
		if group.Name() == name {
			return nil, fmt.Errorf("group named '%s' already exists", name)
		}
	}
	newGroup := g.group.NewSubgroup()
	newGroup.Name = name
	return WrapGroup(newGroup, g.db), nil
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
	for _, each := range subgroup.Groups() {
		if err := subgroup.RemoveSubgroup(each); err != nil {
			return fmt.Errorf("could not purge subgroups in group '%s': %s", each.Name(), err)
		}
	}
	for _, e := range subgroup.Entries() {
		if err := subgroup.RemoveEntry(e); err != nil {
			return fmt.Errorf("could not purge entries in group '%s': %s", e.Title(), err)
		}
	}
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
	// FIXME allows dupe entries
	entry, err := g.group.NewEntry()
	if err != nil {
		return nil, err
	}
	entry.Title = name
	return WrapEntry(entry, g.db), nil
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
