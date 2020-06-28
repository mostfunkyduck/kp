package main

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		if len(args) == 0 {
			// FIXME db.Root() can come in from the context
			currentLocation = getRoot(currentLocation)
		} else {
			newLocation, err := traversePath(currentLocation, args[0])
			if err != nil {
				shell.Println(fmt.Sprintf("invalid path: %s", err))
				return
			}
			currentLocation = newLocation
		}
		changeDirectory(currentLocation, shell)
	}
}

func changeDirectory(newLocation *keepass.Group, shell *ishell.Shell) {
	shell.Set("currentLocation", newLocation)
	shell.SetPrompt(fmt.Sprintf("%s > ", newLocation.Name))
}
