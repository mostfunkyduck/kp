package commands

import (
	"fmt"

	"github.com/mostfunkyduck/ishell"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(t.Database)
		args := c.Args
		currentLocation := db.CurrentLocation()
		if len(c.Args) == 0 {
			currentLocation = db.Root()
		} else {
			newLocation, entry, err := TraversePath(db, currentLocation, args[0])
			if err != nil {
				shell.Println(fmt.Sprintf("invalid path: %s", err))
				return
			}

			if entry != nil {
				shell.Printf("'%s' is an entry, not a group\n", args[0])
				return
			}
			currentLocation = newLocation
		}
		changeDirectory(db, currentLocation, shell)
	}
}

func changeDirectory(db t.Database, newLocation t.Group, shell *ishell.Shell) {
	db.SetCurrentLocation(newLocation)
	path, err := db.Path()
	if err != nil {
		shell.Println("could not render DB path: %s\n", err)
		return
	}
	shell.SetPrompt(fmt.Sprintf("%s > ", path))
}
