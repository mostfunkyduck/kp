package main

import (
	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Pwd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		shell.Println(getPwd(shell, currentLocation))
	}
}
