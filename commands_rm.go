package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// purgeGroup recursively removes all subgroups and entries from a group
func purgeGroup(group *keepass.Group) error {
	for _, e := range group.Entries() {
		if err := group.RemoveEntry(e); err != nil {
			return fmt.Errorf("could not remove entry '%s' from group '%s': %s", e.Title, group.Name, err)
		}
	}
	for _, g := range group.Groups() {
		if err := purgeGroup(g); err != nil {
			return fmt.Errorf("could not purge group %s: %s", g.Name, err)
		}
		if err := group.RemoveSubgroup(g); err != nil {
			return fmt.Errorf("could not remove group %s: %s", g.Name, err)
		}
	}
	return nil
}

func removeEntry(parentLocation *keepass.Group, entryName string) error {
	for i, e := range parentLocation.Entries() {
		if e.Title == entryName || strconv.Itoa(i) == entryName {
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

		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		if currentLocation == nil {
			shell.Println("not at valid location, cannot remove group")
			return
		}

		newLocation, err := traversePath(currentLocation, targetPath)
		if err != nil {
			shell.Printf("could not reach location %s: %s", targetPath, err)
			return
		}

		// trim down to the actual name of the entity we want to kill
		pathbits := strings.Split(targetPath, "/")
		target := pathbits[len(pathbits)-1]
		if groupMode {
			if newLocation.Parent() == nil {
				shell.Println("cannot remove root node")
				return
			}
			if err := purgeGroup(newLocation); err != nil {
				shell.Printf("could not fully remove group '%s': %s\n", newLocation.Name, err)
				return
			}

			if currentLocation == newLocation {
				changeDirectory(currentLocation.Parent(), shell)
			}

			if err := newLocation.Parent().RemoveSubgroup(newLocation); err != nil {
				shell.Printf("could not fully remove group %s: %s\n", newLocation.Name, err)
				return
			}
			return
		}

		if err := removeEntry(newLocation, target); err != nil {
			shell.Printf("error removing entry: %s\n", err)
			return
		}
		shell.Printf("successfully removed '%s'\n", targetPath)
	}
}
