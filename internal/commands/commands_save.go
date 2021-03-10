package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/mostfunkyduck/kp/keepass"
)

func Save(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(keepass.Database)
		savePath := db.SavePath()
		if savePath == "" {
			shell.Println("no path associated with this database! use 'saveas' if this is the first time saving the file")
			return
		}

		if err := db.Save(); err != nil {
			shell.Printf("error saving database: %s\n", err)
			return
		}
		shell.Printf("saved to '%s'\n", savePath)
	}
}
