package keepassv1

import (
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

func (g *Group) Name() string {
	return g.group.Name
}

func (g *Group) Parent() k.Group {
	return NewGroup(g.group.Parent())
}

func (g *Group) Entries() (rv []k.Entry) {
	for _, each := range g.group.Entries() {
		rv = append(rv, NewEntry(each))
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

func (g *Group) Raw() interface{} {
	return g.group
}

func (g *Group) RemoveEntry(e k.Entry) error {
	return g.group.RemoveEntry(e.Raw().(*keepass.Entry))
}
