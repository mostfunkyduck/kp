package main

import (
	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Save(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		filePath := shell.Get("filePath").(string)
		if filePath == "" {
			shell.Println("no path associated with this database! use 'saveas' if this is the first time saving the file")
			return
		}

		db := shell.Get("db").(*keepass.Database)
		if err := saveDB(db, filePath); err != nil {
			shell.Printf("error saving database: %s\n", err)
			return
		}
		shell.Printf("saved to '%s'\n", filePath)
	}
}
