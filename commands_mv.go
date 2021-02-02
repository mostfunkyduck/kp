package main

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func finish(shell *ishell.Shell) {
	DBChanged = true
	if err := promptAndSave(shell); err != nil {
		shell.Printf("error saving database: %s\n", err)
		return
	}
}

func moveEntry(shell *ishell.Shell, e k.Entry, db k.Database, location string) error {
	parent, existingEntry, err := TraversePath(db, db.CurrentLocation(), location)
	if existingEntry != nil {
		shell.Printf("'%s' already exists! overwrite? [y/N]  ")
		input, err := shell.ReadLineErr()
		if err != nil {
			return fmt.Errorf("error reading user input: %s\n", err)
		}

		if input != "y" {
			return fmt.Errorf("not overwriting")
		}

		if err := e.SetParent(existingEntry.Parent()); err != nil {
			return fmt.Errorf("could not move entry '%s' to group '%s': %s\n", string(e.Title()), existingEntry.Parent().Name(), err)
		}

		if err := existingEntry.Parent().RemoveEntry(existingEntry); err != nil {
			return fmt.Errorf("error removing entry '%s' from group '%s': %s\n", existingEntry.Title(), existingEntry.Parent().Name(), err)
		}
		return nil
	}

	title := ""
	if err != nil {
		// there's no group or entry at this location, attempt to process this as a rename
		// trim the path so that we're only looking at the parent group
		pathBits := strings.Split(location, "/")
		path := strings.Join(pathBits[0:len(pathBits)-1], "/")
		var entry k.Entry
		parent, entry, err = TraversePath(db, db.CurrentLocation(), path)
		if err != nil {
			return fmt.Errorf("error finding path '%s': %s\n", location, err)
		}

		if entry != nil {
			return fmt.Errorf("could not rename '%s' to '%s': '%s' is an existing entry", e.Title(), location, path)
		}
		title = pathBits[len(pathBits)-1]
	}

	if err := e.SetParent(parent); err != nil {
		return fmt.Errorf("error moving entry '%s' to new location '%s': %s\n", e.Title(), parent.Name(), err)
	}

	if title != "" {
		e.SetTitle(title)
	}
	return nil
}

func moveGroup(g k.Group, db k.Database, location string) error {
	newNameBits := strings.Split(location, "/")
	newName := newNameBits[len(newNameBits)-1]
	if newName == "" {
		// this should happen if the user moves a group to be a subgroup of another group
		// i.e "mv foo bar/", expecting "foo" to become "bar/foo"
		newName = g.Name()
	}
	newNameParent := strings.Join(newNameBits[0:len(newNameBits)-1], "/")
	parent, _, err := TraversePath(db, db.CurrentLocation(), newNameParent)
	if err != nil {
		return err
	}

	for _, g := range parent.Groups() {
		if g.Name() == newName {
			// attempted to move a group into a group, i.e. '/foo' into '/bar/foo' when '/bar/foo' already exists
			// in this case, we want to move '/foo' to become '/bar/foo/foo', so set the parent to be the target group
			parent = g
			break
		}
	}

	for _, e := range parent.Entries() {
		if e.Title() == newName {
			return fmt.Errorf("entry named '%s' already exists at '%s'", newName, e.Title())
		}
	}

	if err := g.SetParent(parent); err != nil {
		return err
	}
	g.SetName(newName)
	return nil
}

func Mv(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		syntaxCheck(c, 2)
		srcPath := c.Args[0]
		dstPath := c.Args[1]
		db := shell.Get("db").(k.Database)

		l, e, err := TraversePath(db, db.CurrentLocation(), srcPath)
		if err != nil {
			shell.Printf("error parsing path %s: %s", srcPath, err)
			return
		}

		// is this an entry or a group?
		if e != nil {
			if err := moveEntry(shell, e, db, dstPath); err != nil {
				shell.Printf("couldn't move entry: %s", err)
				return
			}
		} else {
			// Not an entry, this is a group
			if err := moveGroup(l, db, dstPath); err != nil {
				shell.Printf("could not move group: %s\n", err)
				return
			}
		}
		finish(shell)
	}
}
