package main

import (
	"regexp"
	"strings"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func NewGroup(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		// We need to remove the final slash as it will confuse the parser
		finalSlashRE := regexp.MustCompile(`/$`)
		newPath := string(finalSlashRE.ReplaceAll([]byte(c.Args[0]), []byte("")))
		if isPresent(shell, newPath) {
			shell.Printf("cannot create duplicate entity '%s'\n", newPath)
			return
		}

		db := shell.Get("db").(k.Database)
		currentLocation := db.CurrentLocation()

		var parentPath string
		// get the parent path of the new group so that we can add the subgroup to it
		parentPathSplit := strings.Split(newPath, "/")
		// rejoin the string components, trimming off the final entry
		parentEntryIndex := len(parentPathSplit) - 1
		if parentEntryIndex <= 0 {
			// this means that there were no parent entries, so the new group will
			// be in the current directory
			parentPath = "."
		} else {
			// get the path of the parent
			parentPath = strings.Join(parentPathSplit[0:parentEntryIndex], "/")
		}

		location, _, err := TraversePath(db, currentLocation, parentPath)
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		re := regexp.MustCompile(`.*/`)
		newGroupName := string(re.ReplaceAll([]byte(newPath), []byte("")))
		l, err := location.NewSubgroup(newGroupName)
		if err != nil {
			shell.Printf("could not create subgroup: %s\n", err)
			return
		}

		p, err := l.Path()
		if err != nil {
			shell.Printf("error getting path: %s\n", p)
		}

		shell.Printf("new location: %s\n", p)

		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("could not save database: %s\n", err)
		}
	}
}
