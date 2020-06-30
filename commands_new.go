package main

import (
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func NewEntry(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		if isPresent(shell, args[0]) {
			shell.Printf("cannot create duplicate entity '%s'\n", args[0])
			return
		}

		currentLocation := shell.Get("currentLocation").(*keepass.Group)

		path := strings.Split(args[0], "/")
		location, err := traversePath(currentLocation, strings.Join(path[0:len(path)-1], "/"))
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		if location.IsRoot() {
			shell.Println("cannot add entries to root node")
			return
		}

		shell.ShowPrompt(false)
		entry, err := location.NewEntry()
		if err != nil {
			shell.Printf("error creating new entry: %s\n", err)
			return
		}
		entry.CreationTime = time.Now()
		entry.LastModificationTime = time.Now()
		entry.LastAccessTime = time.Now()

		err = promptForEntry(shell, entry, path[len(path)-1])
		shell.ShowPrompt(true)
		if err != nil {
			shell.Printf("could not collect user input: %s\n", err)
			if err := location.RemoveEntry(entry); err != nil {
				shell.Printf("could not remove malformed entry from group: %s\n", err)
			}
			return
		}

		if err := promptAndSave(shell); err != nil {
			shell.Printf("failed to save database: %s\n", err)
			return
		}
	}
}
