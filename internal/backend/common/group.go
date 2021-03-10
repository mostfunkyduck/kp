package common

import (
	"fmt"
	"regexp"

	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

type Group struct {
	db     t.Database
	driver t.Group
}

func (g *Group) Path() (rv string, err error) {
	if g.driver.IsRoot() {
		return "/", nil
	}
	pathGroups, err := FindPathToGroup(g.db.Root(), g.driver)
	if err != nil {
		return rv, fmt.Errorf("could not find path to group '%s'", g.driver.Name())
	}
	for _, each := range pathGroups {
		rv = rv + each.Name() + "/"
	}
	return rv + g.driver.Name() + "/", nil
}

func FindPathToGroup(source t.Group, target t.Group) (rv []t.Group, err error) {
	// the v2 library doesn't appear to support child->parent links, so we have to find the needful ourselves

	// loop through every group in the top level of the path
	for _, group := range source.Groups() {
		same, err := CompareUUIDs(group, target)
		if err != nil {
			return []t.Group{}, fmt.Errorf("could not compare UUIDS: %s", err)
		}

		// If the group that we're looking at in the path is the target,
		// then the 'source' group at the top level is part of the final path
		if same {
			// this will essentially say that if the target group exists in the source group, build a path
			// to the source group
			ret := []t.Group{source}
			return ret, nil
		}

		// If the group is not the exact match, recurse into it looking for the target
		// If the target is in a subgroup tree here, the full path will end up being returned
		pathGroups, err := FindPathToGroup(group, target)
		if err != nil {
			return []t.Group{}, fmt.Errorf("could not find path from group '%s' to group '%s': %s", group.Name(), target.Name(), err)
		}

		// if the target group is a child of this group, return the full path
		if len(pathGroups) != 0 {
			ret := append([]t.Group{source}, pathGroups...)
			return ret, nil
		}
	}
	return []t.Group{}, nil
}

func (g *Group) DB() t.Database {
	return g.db
}

func (g *Group) SetDB(d t.Database) {
	g.db = d
}

// sets pointer to the version of itself that can access child methods... FIXME this is a bit of a mind bender
func (g *Group) SetDriver(gr t.Group) {
	g.driver = gr
}

func (g *Group) Search(term *regexp.Regexp) (paths []string, err error) {
	if term.FindString(g.driver.Name()) != "" {
		path, err := g.Path()
		if err == nil {
			// append slash so it's clear that it's a group, not an entry
			paths = append(paths, path)
		}
	}

	for _, e := range g.driver.Entries() {
		nestedSearch, err := e.Search(term)
		if err != nil {
			return []string{}, fmt.Errorf("search failed on entries: %s", err)
		}
		paths = append(paths, nestedSearch...)
	}

	for _, g := range g.driver.Groups() {
		nestedSearch, err := g.Search(term)
		if err != nil {
			return []string{}, fmt.Errorf("search failed while recursing into groups: %s", err)
		}
		paths = append(paths, nestedSearch...)
	}
	return paths, nil
}
