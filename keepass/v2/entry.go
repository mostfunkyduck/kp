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
	entry g.Entry
}

// findPathToEntry returns all the groups in the path leading to an entry *but not the entry itself*
func findPathToEntry(source k.Group, target k.Entry) (rv []k.Group) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, entry := range source.Entries() {
		uuidString, err := target.UUIDString()
		if err != nil {
			// TODO swallowing this error
			return []k.Group{}
		}
		entryUUIDString, err := entry.UUIDString()
		if err != nil {
			return []k.Group{}
		}

		if entryUUIDString == uuidString {
			return []k.Group{source}
		}
	}
	for _, group := range source.Groups() {
		if pathGroups := findPathToEntry(group, target); len(pathGroups) != 0 {
			return append([]k.Group{group}, pathGroups...)
		}
	}
	return []k.Group{}
}

func newEntry(entry g.Entry, db k.Database) k.Entry {
	return &Entry{
		db:    db,
		entry: entry,
	}
}

func (e *Entry) Raw() interface{} {
	return e.entry
}

func (e *Entry) Path() (path string) {
	pathGroups := findPathToEntry(e.db.Root(), e)
	for _, each := range pathGroups {
		path = path + "/" + each.Name()
	}
	path = path + "/"
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
	return false
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
	pathGroups := findPathToEntry(e.db.Root(), e)
	if len(pathGroups) == 0 {
		return nil
	}

	return pathGroups[len(pathGroups)-1]
}

func (e *Entry) SetParent(g k.Group) error {
	pathGroups := findPathToGroup(e.db.Root(), g)
	if len(pathGroups) == 0 {
		return fmt.Errorf("could not find a path from the db root to '%s', is this a valid group?", g.Name())
	}
	parent := pathGroups[len(pathGroups)-1]
	for _, each := range parent.Entries() {
		if e.GetTitle() == each.GetTitle() {
			// no dupes!
			return fmt.Errorf("could not change parent: duplicate entry exists at target location")
		}
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

func (e *Entry) GetPassword() string {
	return e.entry.GetPassword()
}

func (e *Entry) GetTitle() string {
	return e.entry.GetTitle()
}
