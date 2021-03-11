package commands

import (
	"github.com/abiosoft/ishell"
)

func Xu(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		if err := copyFromEntry(shell, c.Args[0], "username"); err != nil {
			shell.Printf("could not copy username: %s", err)
			return
		}
	}
}
