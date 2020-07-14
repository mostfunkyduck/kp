package main

import (
	"strings"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
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
				shell.Printf("could move entry '%s' to group '%s': %s\n", srcEntry.Get("title").Value().(string), existingEntry.Parent().Name, err)
				return
			}
			if err := existingEntry.Parent().RemoveEntry(existingEntry); err != nil {
				shell.Printf("error removing entry '%s' from group '%s': %s\n", existingEntry.Get("title").Value(), existingEntry.Parent().Name, err)
				return
			}
			return
		}

		title := ""
		db := shell.Get("db").(k.Database)
		location, err := db.TraversePath(db.CurrentLocation(), dstPath)
		if err != nil {
			// there's no group or entry at this location, attempt to process this as a rename
			// and set the location to be the group
			pathBits := strings.Split(dstPath, "/")
			path := strings.Join(pathBits[0:len(pathBits)-1], "/")
			location, err = db.TraversePath(db.CurrentLocation(), path)
			if err != nil {
				shell.Printf("error finding path '%s': %s\n", dstPath, err)
				return
			}
			title = pathBits[len(pathBits)-1]
		}

		if err := srcEntry.SetParent(location); err != nil {
			shell.Printf("error moving entry '%s' to new location '%s': %s\n", srcEntry.Get("title").Value().(string), location.Name, err)
			return
		}
		if title != "" {
			srcEntry.Set("title", title)
		}

		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("error saving database: %s\n", err)
			return
		}
	}
}
