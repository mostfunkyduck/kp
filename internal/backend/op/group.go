package op

import (
	"fmt"
	"regexp"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

type Group struct{
	c.Group
	UUID	string
	Title	string
	opDriver *Driver
}

func NewGroup(title string, db t.Database, d *Driver) Group{
	g := Group{
		Title: title,
	}
	g.SetDB(db)
	g.opDriver = d
	return g
}

func (g Group) Raw() interface{} {
	return g
}

// Path returns a pretty simple interpretation as the path because op doesn't have nested vaults
func (g Group) Path() (string, error) {
	return "/" + g.Title, nil
}

// Search searches this object and all nested objects for a given regular expression
func (g Group) Search(r *regexp.Regexp) ([]string, error) {
	return []string{}, fmt.Errorf("not implemented")
}

func (g Group) UUIDString() (string, error) {
	return g.UUID, nil
}

func (g Group) Entries() []t.Entry {
	if g.IsRoot() {
		return []t.Entry{}
	}
	entries, err := g.opDriver.ListItems(g.Title)
	if err != nil {
		return []t.Entry{}
	}
	ret := make([]t.Entry, len(entries))
	for i, v := range entries {
		ret[i] = t.Entry(v) // what have i done, idk why this is necessary
	}
	return ret
}

// the other backends have to use WrapGroup, which isn't neccessary here, so...
func stupidWorkaround(groups []Group, d *Driver) []t.Group {
	// this hurts my soul https://stackoverflow.com/questions/12994679/slice-of-struct-slice-of-interface-it-implements
	ret := make([]t.Group, len(groups))
	for i, v := range groups {
		ret[i] = t.Group(NewGroup(v.Title, v.DB(), d))
	}
	return ret
}
func (g Group) Groups() []t.Group {
	if !g.IsRoot() {
		return []t.Group{}
	}
	groups, err := g.opDriver.ListVaults()
	if err != nil {
		fmt.Printf("error encountered retrieving vaults, swallowing it, yummy yummy: %s\n", err)
		return []t.Group{}
	}
	return stupidWorkaround(groups, g.opDriver)
}

func (g Group) Parent() t.Group {
	if !g.IsRoot() {
		return g.DB().Root()
	}
	return nil
}

func (g Group) SetParent(t.Group) error {
	return fmt.Errorf("cannot set parent on op group, not supported by 1password")
}

func (g Group) AddEntry(t.Entry) error {
	return fmt.Errorf("not implemented")
}

func (g Group) Name() string {
	return g.Title
}

// no-op until this is supported
func (g Group) SetName(string) {
	return
}

func (g Group) IsRoot() bool {
	return g.Title == ""
}

func (g Group) NewSubgroup(name string) (t.Group, error) {
	if !g.IsRoot() {
		return nil, fmt.Errorf("not supported by 1password")
	}
	return nil, fmt.Errorf("not supported")
}
func (g Group) RemoveSubgroup(t.Group) error {
	return fmt.Errorf("not supported")
}
func (g Group) AddSubgroup(t.Group) error {
	if !g.IsRoot() {
		return fmt.Errorf("not supported by 1password")
	}
	return fmt.Errorf("not supported")
}
func (g Group) NewEntry(name string) (t.Entry, error) {
	return nil, fmt.Errorf("not supported")
}
func (g Group) RemoveEntry(t.Entry) error {
	return fmt.Errorf("not supported")
}
