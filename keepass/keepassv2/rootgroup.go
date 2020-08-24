package keepassv2

// Unlike the keepass 1 library, this library doesn't represent Root as a group
// which means that we have to dress up its 'RootData' object as a Group object

import (
	"fmt"
	"regexp"

	k "github.com/mostfunkyduck/kp/keepass"
	g "github.com/tobischo/gokeepasslib/v3"
)

type RootGroup struct {
	db   k.Database
	root *g.RootData
}

func (r *RootGroup) Raw() interface{} {
	return r.root
}

func (r *RootGroup) Groups() (rv []k.Group) {
	for i := range r.root.Groups {
		rv = append(rv, WrapGroup(&r.root.Groups[i], r.db))
	}
	return
}

func (r *RootGroup) Path() (string, error) {
	return "/", nil
}

// technically, this could return all the entries in the database, but since
// that's inconsistent with other groups, leaving it this way for now
func (r *RootGroup) Entries() (rv []k.Entry) {
	return []k.Entry{}
}

func (r *RootGroup) Parent() k.Group {
	return nil
}

func (r *RootGroup) SetParent(parent k.Group) error {
	return fmt.Errorf("cannot set parent for root group")
}

func (r *RootGroup) Name() string {
	return ""
}

func (r *RootGroup) SetName(name string) {
}

func (r *RootGroup) IsRoot() bool {
	return true
}

// Creates a new subgroup with a given name under this group
func (r *RootGroup) NewSubgroup(name string) (k.Group, error) {
	newGroup := g.NewGroup()
	newGroupWrapper := WrapGroup(&newGroup, r.db)
	newGroupWrapper.SetName(name)
	if err := newGroupWrapper.SetParent(r); err != nil {
		return &Group{}, fmt.Errorf("couldn't assign new group to parent '%s'; %s", r.Name(), err)
	}
	return newGroupWrapper, nil
}

func (r *RootGroup) RemoveSubgroup(subgroup k.Group) error {
	subUUID, err := subgroup.UUIDString()
	if err != nil {
		return fmt.Errorf("could not read UUID on '%s': %s", subgroup.Name(), err)
	}

	for i, each := range r.root.Groups {
		eachWrapper := WrapGroup(&each, r.db)
		eachUUID, err := eachWrapper.UUIDString()
		if err != nil {
			return fmt.Errorf("could not read UUID on '%s': %s", eachWrapper.Name(), err)
		}

		if eachUUID == subUUID {
			// remove it
			raw := r.root
			groupLen := len(raw.Groups)
			raw.Groups = append(raw.Groups[0:i], raw.Groups[i+1:groupLen]...)
			return nil
		}
	}
	return fmt.Errorf("could not find group with UUID '%s'", subUUID)
}

func (r *RootGroup) AddEntry(e k.Entry) error {
	return fmt.Errorf("cannot add entries to root group")
}
func (r *RootGroup) NewEntry(name string) (k.Entry, error) {
	return nil, fmt.Errorf("cannot add entries to root group")
}

func (r *RootGroup) RemoveEntry(entry k.Entry) error {
	return fmt.Errorf("root group does not hold entries")
}

func (r *RootGroup) Search(term *regexp.Regexp) (paths []string) {
	for _, g := range r.Groups() {
		paths = append(paths, g.Search(term)...)
	}
	return paths
}

func (r *RootGroup) UUIDString() (string, error) {
	return "<root group>", nil
}

func (r *RootGroup) AddSubgroup(subgroup k.Group) error {
	for _, each := range r.Groups() {
		if each.Name() == subgroup.Name() {
			return fmt.Errorf("group named '%s' already exists", each.Name())
		}
	}

	// FIXME this pointer abomination needs to go
	r.root.Groups = append(r.root.Groups, *subgroup.Raw().(*g.Group))
	subgroup.(*Group).updateWrapper(&r.root.Groups[len(r.root.Groups)-1])
	return nil
}
