package main

import (
	"fmt"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func Ls(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(k.Database)
		currentLocation := db.CurrentLocation()
		location := currentLocation
		if len(c.Args) > 0 {
			newLocation, entry, err := TraversePath(db, currentLocation, c.Args[0])
			if err != nil {
				shell.Printf("invalid path: %s\n", err)
				return
			}

			// if this is the path to an entry, just output that and be done with it
			if entry != nil {
				shell.Printf("%s\n", entry.Get("title").Value.(string))
				return
			}

			location = newLocation
		}

		lines := []string{}
		lines = append(lines, "=== Groups ===")
		for _, group := range location.Groups() {
			lines = append(lines, fmt.Sprintf("%s/", group.Name()))
		}

		lines = append(lines, "\n=== Entries ===")
		for i, entry := range location.Entries() {
			lines = append(lines, fmt.Sprintf("%d: %s", i, entry.Get("title").Value.(string)))
		}
		for _, line := range lines {
			shell.Println(line)
		}
	}
}
