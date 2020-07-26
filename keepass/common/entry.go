package common

// Entry is a wrapper around an entry driver, holding functions
// common to both kp1 and kp2
import (
	"fmt"
	"regexp"
	"strings"
	k "github.com/mostfunkyduck/kp/keepass"
)

type Entry struct {
	db k.Database
	entry k.Entry
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

// Path returns the fully qualified path to the entry, if there's no parent, only the name is returned
func (e *Entry) Path() (path string, err error) {
	pathGroups, err := findPathToEntry(e.DB().Root(), e.entry)
	if err != nil {
		return path, fmt.Errorf("could not find path from root to %s: %s", e.entry.Title(), err)
	}

	if len(pathGroups) > 0 {
		// if we're about to build a path, start it off properly
		// otherwise, this string will stay empty until the title is inserted at the end
		path = "/"
	}
	for _, each := range pathGroups {
		path = path + each.Name() + "/"
	}
	path = path + e.entry.Title()
	return
}

func (e *Entry) Parent() k.Group {
	pathGroups, err := findPathToEntry(e.DB().Root(), e.entry)
	if err != nil {
		return nil
	}
	if len(pathGroups) == 0 {
		return nil
	}

	return pathGroups[len(pathGroups)-1]
}

func (e *Entry) SetParent(g k.Group) error {
	pathGroups, err := findPathToGroup(e.DB().Root(), g)
	if len(pathGroups) == 0 || err != nil {
		errorString := fmt.Sprintf("could not find a path from the db root to '%s', is this a valid group?", g.Name())

		if err != nil {
			errorString = errorString + fmt.Sprintf(" (error occurred: %s)", err)
		}
		return fmt.Errorf(errorString)
	}

	if err := g.AddEntry(e.entry); err != nil {
		return fmt.Errorf("cannot add entry to group: %s", err)
	}
	return nil
}

func (e *Entry) Output(full bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "=== Values ===\n")
	fmt.Fprintf(&b, "index\tkey\tvalue\tprotected\n")
	for idx, val := range e.entry.Values() {
		fmt.Fprintf(&b, "%d\t|\t%s\t|\t%s\t|\t%t\n", idx, val.Name, val.Value.(string), val.Protected)
	}
	return b.String()
}

// TODO test various fields to make sure they are searchable, consider adding searchability toggle
func (e *Entry) Search(term *regexp.Regexp) (paths []string) {
	for _, val := range e.entry.Values() {
		content := val.Value.(string)
		if term.FindString(content) != "" ||
			term.FindString(val.Name) != "" {
			// something in this entry matched, let's return it
			path, _ := e.Path()
			paths = append(paths, path)
			break
		}
	}

	return
}

func (e *Entry) DB() k.Database {
	return e.db
}

func (e *Entry) SetDB(db k.Database) {
	e.db = db
}

// SetEntry sets the internal entry driver for this wrapper
func (e *Entry) SetEntry(entry k.Entry) {
	e.entry = entry
}