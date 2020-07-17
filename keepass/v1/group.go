package keepassv1

import (
	"fmt"
	"regexp"
	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Group struct {
	group *keepass.Group
}

func NewGroup(group *keepass.Group) k.Group {
	return &Group{
		group: group,
	}
}


// FIXME the keepass library has a bug where you can't get the parent
// unless the entry is a pointer to the one in the db (it's comparing pointer values)
// this can/should/will be fixed in my fork
func (g *Group) searchEntries(term *regexp.Regexp) (titles []string) {
	for _, e := range g.Entries() {
		if term.FindString(e.Get("title").Value.(string)) != "" ||
			term.FindString(e.Get("notes").Value.(string)) != "" ||
			term.FindString(e.Get("attachment").Name) != "" ||
			term.FindString(e.Get("username").Value.(string)) != "" {
			titles = append(titles, e.Get("title").Value.(string))
		}
	}
	return titles
}

func (g *Group) Search(term *regexp.Regexp) (paths []string) {

	for _, title := range g.searchEntries(term) {
		paths = append(paths, "./"+title)
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
	return NewGroup(g.group.Parent())
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
		rv = append(rv, NewGroup(each))
	}
	return rv
}

func (g *Group) IsRoot() bool {
	return g.Parent() == nil
}

func (g *Group) NewSubgroup(name string) k.Group {
	newGroup := g.group.NewSubgroup()
	newGroup.Name = name
	return &Group{
		group: newGroup,
	}
}

func (g *Group) RemoveSubgroup(subgroup k.Group) error {
	return g.group.RemoveSubgroup(subgroup.Raw().(*keepass.Group))
}

func (g *Group) Path() (fullPath string) {
	group := g.group
	for ; group != nil; group = group.Parent() {
		if group.IsRoot() {
			fullPath = "/" + fullPath
			break
		}
		fullPath = group.Name + "/" + fullPath
	}
	return fullPath
}

func (g *Group) Raw() interface{} {
	return g.group
}

func (g *Group) NewEntry() (k.Entry, error) {
	entry, err := g.group.NewEntry()
	if err != nil {
		return nil, err
	}
	return &Entry{
		entry: entry,
	}, nil
}

func (g *Group) RemoveEntry(e k.Entry) error {
	return g.group.RemoveEntry(e.Raw().(*keepass.Entry))
}
