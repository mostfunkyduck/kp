package commands

import (
	"github.com/mostfunkyduck/ishell"
)

func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		fullMode := false
		path := buildPath(c.Args)

		for _, flag := range c.Flags {
			if flag == "-f" {
				fullMode = true
				path = buildPath(c.Args[1:])
			}
		}

		entry, ok := getEntryByPath(shell, path)
		if !ok {
			shell.Printf("could not retrieve entry at path '%s'\n", path)
			return
		}

		shell.Println(entry.Output(fullMode))
	}
}
