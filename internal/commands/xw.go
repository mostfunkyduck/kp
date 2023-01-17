package commands

import (
	"github.com/mostfunkyduck/ishell"
)

func Xw(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		path := buildPath(c.Args)
		if !ok {
			shell.Println(errString)
			return
		}

		if err := copyFromEntry(shell, path, "url"); err != nil {
			shell.Printf("could not copy url: %s", err)
			return
		}
	}
}
