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
		// the user may enter a path, either absolute or relative, in which case
		// we need to crawl to that location to make the new entry

		// If the path passed in has a slash, that means we need to crawl, otherwise
		// just stay put by 'crawling' to '.'
		targetPath := "."

		// if the path doesn't have a slash in it, then it represents the group name
		groupName := newPath
		if strings.Contains(newPath, "/") {
			// to reach the correct location, trim off everything after
			// the last slash and crawl to the path represented by what's left over
			r := regexp.MustCompile(`(?P<Path>.*)/(?P<Group>[^/]*)$`)
			matches := r.FindStringSubmatch(newPath)
			targetPath = string(matches[1])
			// save off the group name for later use
			groupName = string(matches[2])
		}

		// use TraversePath to crawl to the target path
		location, _, err := TraversePath(db, db.CurrentLocation(), targetPath)
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		l, err := location.NewSubgroup(groupName)
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
