package main

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Ls(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location := currentLocation
		entityName := "/"
		if len(c.Args) > 0 {
			path := strings.Split(c.Args[0], "/")
			entityName = path[len(path)-1]
			newLocation, err := traversePath(currentLocation, c.Args[0])
			if err != nil {
				shell.Printf("Invalid path: %s", err)
				return
			}
			location = newLocation
		}

		lines := []string{}
		lines = append(lines, "=== Groups ===")
		for _, group := range location.Groups() {
			if group.Name == entityName {
				shell.Println(group.Name + "/")
				return
			}
			lines = append(lines, fmt.Sprintf("%s/", group.Name))
		}

		lines = append(lines, "\n=== Entries ===")
		for i, entry := range location.Entries() {
			// entryLine := fmt.Sprintf("%d: %s", i, entry.Title)
			lines = append(lines, fmt.Sprintf("%d: %s", i, entry.Title))
			if entry.Title == entityName {
				shell.Println(entry.Title)
				return
			}
		}
		for _, line := range lines {
			shell.Println(line)
		}
	}
}
