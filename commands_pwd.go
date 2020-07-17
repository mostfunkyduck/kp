package main

import (
	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

func Pwd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(k.Database)
		shell.Println(db.Path())
	}
}
