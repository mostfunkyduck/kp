package main

import (
	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Mv(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		syntaxCheck(c, 2)
		srcPath := c.Args[0]
		dstPath := c.Args[1]
		srcEntry, ok := getEntryByPath(shell, srcPath)
		if !ok {
			shell.Printf("invalid source path: %s\n", srcPath)
			return
		}

		existingEntry, ok := getEntryByPath(shell, dstPath)
		if ok {
			shell.Printf("'%s' already exists! overwrite? [y/N]  ")
			input, err := shell.ReadLineErr()
			if err != nil {
				shell.Printf("error reading user input: %s\n", err)
				return
			}

			if input != "y" {
				shell.Println("not overwriting")
				return
			}
			if err := srcEntry.SetParent(existingEntry.Parent()); err != nil {
				shell.Printf("could move entry '%s' to group '%s': %s\n", srcEntry.Title, existingEntry.Parent().Name, err)
				return
			}
			if err := existingEntry.Parent().RemoveEntry(existingEntry); err != nil {
				shell.Printf("error removing entry '%s' from group '%s': %s\n", existingEntry.Title, existingEntry.Parent().Name, err)
				return
			}
			return
		}

		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, dstPath)
		if err != nil {
			shell.Printf("error finding path '%s': %s\n", dstPath, err)
			return
		}

		if err := srcEntry.SetParent(location); err != nil {
			shell.Printf("error moving entry '%s' to new location '%s': %s\n", srcEntry.Title, location.Name, err)
			return
		}
		if err := promptAndSave(shell); err != nil {
			shell.Printf("error saving database: %s\n", err)
			return
		}
	}
}
