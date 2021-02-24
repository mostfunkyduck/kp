package main

import (
	"strings"

	"github.com/abiosoft/ishell"
)

func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		fullMode := false
		path := c.Args[0]
		for _, arg := range c.Args {
			if strings.HasPrefix(arg, "-") {
				if arg == "-f" {
					fullMode = true
				}
				continue
			}
			path = arg
		}

		entry, ok := getEntryByPath(shell, path)
		if !ok {
			shell.Printf("could not retrieve entry at path '%s'\n", path)
			return
		}

		shell.Println(entry.Output(fullMode))
	}
}
