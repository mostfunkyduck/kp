package common

import (
	"fmt"
	"regexp"

	k "github.com/mostfunkyduck/kp/keepass"
)

type Group struct {
	db     k.Database
	driver k.Group
}

func (g *Group) Path() (rv string, err error) {
	if g.driver.IsRoot() {
		return "/", nil
	}
	pathGroups, err := findPathToGroup(g.db.Root(), g.driver)
	if err != nil {
		return rv, fmt.Errorf("could not find path to group '%s'", g.driver.Name())
	}
	for _, each := range pathGroups {
		rv = rv + each.Name() + "/"
	}
	return rv + g.driver.Name() + "/", nil
}

func findPathToGroup(source k.Group, target k.Group) (rv []k.Group, err error) {
	// the v2 library doesn't appear to support child->parent links, so we have to find the needful ourselves
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

func (g *Group) DB() k.Database {
	return g.db
}

func (g *Group) SetDB(d k.Database) {
	g.db = d
}

// sets pointer to the version of itself that can access child methods... FIXME this is a bit of a mind bender
func (g *Group) SetDriver(gr k.Group) {
	g.driver = gr
}

func (g *Group) Search(term *regexp.Regexp) (paths []string) {
	if term.FindString(g.driver.Name()) != "" {
		path, err := g.Path()
		if err == nil {
			// append slash so it's clear that it's a group, not an entry
			paths = append(paths, path)
		}
	}

	for _, e := range g.driver.Entries() {
		paths = append(paths, e.Search(term)...)
	}

	for _, g := range g.driver.Groups() {
		paths = append(paths, g.Search(term)...)
	}
	return paths
}
