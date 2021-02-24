package main

import (
	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func Pwd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(k.Database)
		path, err := db.Path()
		if err != nil {
			shell.Printf("could not retrieve current path: %s\n", err)
		}
		shell.Println(path)
	}
}
