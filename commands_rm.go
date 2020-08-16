package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

// purgeGroup recursively removes all subgroups and entries from a group
func purgeGroup(group k.Group) error {
	for _, e := range group.Entries() {
		if err := group.RemoveEntry(e); err != nil {
			return fmt.Errorf("could not remove entry '%s' from group '%s': %s", e.Title(), group.Name(), err)
		}
	}
	for _, g := range group.Groups() {
		if err := purgeGroup(g); err != nil {
			return fmt.Errorf("could not purge group %s: %s", g.Name(), err)
		}
		if err := group.RemoveSubgroup(g); err != nil {
			return fmt.Errorf("could not remove group %s: %s", g.Name(), err)
		}
	}
	return nil
}

func removeEntry(parentLocation k.Group, entryName string) error {
	for i, e := range parentLocation.Entries() {
		if e.Title() == entryName || strconv.Itoa(i) == entryName {
			if err := parentLocation.RemoveEntry(e); err != nil {
				return fmt.Errorf("could not remove entry: %s", err)
			}
			return nil
		}
	}
	return fmt.Errorf("could not find entry named '%s'", entryName)
}

func Rm(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		targetPath := c.Args[0]
		groupMode := false
		if len(c.Args) > 1 && c.Args[0] == "-r" {
			groupMode = true
			targetPath = c.Args[1]
		}

		db := shell.Get("db").(k.Database)
		currentLocation := db.CurrentLocation()
		newLocation, entry, err := TraversePath(db, currentLocation, targetPath)
		if err != nil {
			shell.Printf("could not reach location %s: %s", targetPath, err)
			return
		}

		// trim down to the actual name of the entity we want to kill
		pathbits := strings.Split(targetPath, "/")
		target := pathbits[len(pathbits)-1]

		// only remove groups if the specified target was a group
		if entry != nil {
			if err := removeEntry(newLocation, target); err != nil {
				shell.Printf("error removing entry: %s\n", err)
				return
			}
		} else if groupMode {
			if newLocation.Parent() == nil {
				shell.Println("cannot remove root node")
				return
			}
			if err := purgeGroup(newLocation); err != nil {
				shell.Printf("could not fully remove group '%s': %s\n", newLocation.Name, err)
				return
			}

			if currentLocation == newLocation {
				changeDirectory(db, currentLocation.Parent(), shell)
			}

			if err := newLocation.Parent().RemoveSubgroup(newLocation); err != nil {
				shell.Printf("could not fully remove group %s: %s\n", newLocation.Name, err)
				return
			}
			return
		} else {
			shell.Printf("'%s' is a group - try rerunning with '-r'\n", targetPath)
			return
		}

		shell.Printf("successfully removed '%s'\n", targetPath)

		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("could not save: %s\n", err)
		}
	}
}
