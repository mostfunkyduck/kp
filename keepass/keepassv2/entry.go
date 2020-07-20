package keepassv2

import (
	"encoding/base64"
	"fmt"
	k "github.com/mostfunkyduck/kp/keepass"
	g "github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
	"strings"
	"time"
)

type Entry struct {
	db    k.Database
	entry *g.Entry
}

// findPathToEntry returns all the groups in the path leading to an entry *but not the entry itself*
// the path returned will also not include the source group
func findPathToEntry(source k.Group, target k.Entry) (rv []k.Group, err error) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, entry := range source.Entries() {
		uuidString, err := target.UUIDString()
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not get UUID for target '%s': %s", target.Title(), err)
		}
		entryUUIDString, err := entry.UUIDString()
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not get UUID for entry '%s': %s", entry.Title(), err)
		}

		if entryUUIDString == uuidString {
			return []k.Group{source}, nil
		}
	}
	groups := source.Groups()
	for _, group := range groups {
		newGroups, err := findPathToEntry(group, target)
		if err != nil {
			// not putting the path in this error message because it might trigger an infinite loop
			// since this is part of the path traversal algo
			return []k.Group{}, fmt.Errorf("error finding path to '%s' from '%s': %s", target.Title(), group.Name(), err)
		}
		if len(newGroups) > 0 {
			return newGroups, nil
		}
	}
	return []k.Group{}, nil
}

func WrapEntry(entry *g.Entry, db k.Database) k.Entry {
	return &Entry{
		db:    db,
		entry: entry,
	}
}

func (e *Entry) Raw() interface{} {
	return e.entry
}

// returns the fully qualified path to the entry, if there's no parent, only the name is returned
func (e *Entry) Path() (path string, err error) {
	pathGroups, err := findPathToEntry(e.db.Root(), e)
	if err != nil {
		return path, fmt.Errorf("could not find path from root to %s: %s", e.Title(), err)
	}

	if len(pathGroups) > 0 {
		// if we're about to build a path, start it off properly
		// otherwise, this string will stay empty until the title is inserted at the end
		path = "/"
	}
	for _, each := range pathGroups {
		path = path + each.Name() + "/"
	}
	path = path + e.Title()
	return
}

func (e *Entry) UUIDString() (string, error) {
	encodedUUID, err := e.entry.UUID.MarshalText()
	if err != nil {
		return "", fmt.Errorf("could not encode UUID: %s", err)
	}
	str, err := base64.StdEncoding.DecodeString(string(encodedUUID))
	if err != nil {
		return "", fmt.Errorf("could not decode b64: %s", err)
	}
	return string(str), nil
}

func (e Entry) Get(field string) k.Value {
	val := e.entry.Get(field)
	if val == nil {
		return k.Value{}
	}

	return k.Value{
		Name:  field,
		Value: val.Value.Content,
	}
}

func (e *Entry) Set(value k.Value) bool {
	for _, each := range e.entry.Values {
		if each.Key == value.Name {
			oldContent := each.Value.Content
			oldProtected := each.Value.Protected

			// TODO filter for binaries here, bad shit will happen if you try to attach this way :D
			each.Value.Content = value.Value.(string)
			each.Value.Protected = w.NewBoolWrapper(value.Protected)

			return (oldContent != value.Value) || (oldProtected.Bool != value.Protected)
		}
	}
	// no existing value to update, create it fresh
	e.entry.Values = append(e.entry.Values, g.ValueData{
		Key: value.Name,
		Value: g.V{
			Content:   value.Value.(string),
			Protected: w.NewBoolWrapper(value.Protected),
		},
	})
	return true
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.Times.LastAccessTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) SetLastModificationTime(t time.Time) {
	e.entry.Times.LastModificationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) SetCreationTime(t time.Time) {
	e.entry.Times.CreationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) Parent() k.Group {
	pathGroups, err := findPathToEntry(e.db.Root(), e)
	if err != nil {
		return nil
	}
	if len(pathGroups) == 0 {
		return nil
	}

	return pathGroups[len(pathGroups)-1]
}

func (e *Entry) SetParent(g k.Group) error {
	pathGroups, err := findPathToGroup(e.db.Root(), g)
	if len(pathGroups) == 0 || err != nil {
		errorString := fmt.Sprintf("could not find a path from the db root to '%s', is this a valid group?", g.Name())

		if err != nil {
			errorString = errorString + fmt.Sprintf(" (error occurred: %s)", err)
		}
		return fmt.Errorf(errorString)
	}
	parent := pathGroups[len(pathGroups)-1]
	for _, each := range parent.Entries() {
		if e.Title() == each.Title() {
			// no dupes!
			return fmt.Errorf("could not change parent: duplicate entry exists at target location")
		}
	}
	if err := g.AddEntry(e); err != nil {
		return fmt.Errorf("cannot add entry to group: %s", err)
	}
	return nil
}

func (e *Entry) Output(full bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "=== Values ===\n")
	fmt.Fprintf(&b, "index\tkey\tvalue\tprotected\n")
	for idx, val := range e.entry.Values {
		fmt.Fprintf(&b, "%d\t|\t%s\t|\t%s\t|\t%t\n", idx, val.Key, val.Value.Content, val.Value.Protected)
	}
	return b.String()
}

func (e *Entry) Values() (values []k.Value) {
	for _, each := range e.entry.Values {
		newValue := k.Value{
			Name: each.Key,
			Value: each.Value.Content,
			Protected: each.Value.Protected.Bool,
		}
		values = append(values, newValue)
	}
	return
}
func (e *Entry) Password() string {
	return e.entry.GetPassword()
}

func (e *Entry) Title() string {
	return e.entry.GetTitle()
}
