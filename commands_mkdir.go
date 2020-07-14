package main

import (
	"strings"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func NewGroup(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		if isPresent(shell, c.Args[0]) {
			shell.Printf("cannot create duplicate entity '%s'\n", c.Args[0])
			return
		}
		path := strings.Split(c.Args[0], "/")
		db := shell.Get("db").(k.Database)
		location, err := db.TraversePath(db.CurrentLocation(), strings.Join(path[0:len(path)-1], "/"))
		if err != nil {
			shell.Printf("invalid path: " + err.Error())
			return
		}

		location.NewSubgroup(path[len(path)-1])
		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("could not save database: %s\n", err)
		}
	}
}
