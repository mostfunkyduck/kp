package main

import (
	"fmt"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(k.Database)
		args := c.Args
		currentLocation := db.CurrentLocation()
		if len(c.Args) == 0 {
			currentLocation = db.Root()
		}

		newLocation, err := db.TraversePath(currentLocation, args[0])
		if err != nil {
			shell.Println(fmt.Sprintf("invalid path: %s", err))
			return
		}
		changeDirectory(newLocation, shell)
		db.SetCurrentLocation(newLocation)
	}
}

func changeDirectory(newLocation k.Group, shell *ishell.Shell) {
	shell.Set("currentLocation", newLocation)
	shell.SetPrompt(fmt.Sprintf("%s > ", newLocation.Name()))
}
