package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/atotto/clipboard"
)

func Xx(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if err := clipboard.WriteAll(""); err != nil {
			shell.Println("could not clear password from clipboard")
			return
		}
		shell.Println("clipboard cleared!")
	}
}
