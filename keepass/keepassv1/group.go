package keepassv1

import (
	"fmt"

	k "github.com/mostfunkyduck/kp/keepass"
	c "github.com/mostfunkyduck/kp/keepass/common"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Group struct {
	c.Group
	group *keepass.Group
}

func WrapGroup(group *keepass.Group, db k.Database) k.Group {
	if group == nil {
		return nil
	}
	g := &Group{
		group: group,
	}

	g.SetDB(db)
	g.SetDriver(g)
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

func (g *Group) Name() string {
	if g.IsRoot() {
		return ""
	}
	return g.group.Name
}

func (g *Group) SetName(name string) {
	g.group.Name = name
}

func (g *Group) Parent() k.Group {
	return WrapGroup(g.group.Parent(), g.DB())
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
		rv = append(rv, WrapGroup(each, g.DB()))
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
	return WrapGroup(newGroup, g.DB()), nil
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
	return WrapEntry(entry, g.DB()), nil
}

func (g *Group) RemoveEntry(e k.Entry) error {
	return g.group.RemoveEntry(e.Raw().(*keepass.Entry))
}

func (g *Group) UUIDString() (string, error) {
	return string(g.group.ID), nil
}
