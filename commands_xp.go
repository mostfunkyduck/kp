package main

import (
	"github.com/abiosoft/ishell"
	"github.com/atotto/clipboard"
)

func Xp(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		targetPath := c.Args[0]
		entry, ok := getEntryByPath(shell, targetPath)
		if !ok {
			shell.Printf("could not retrieve entry at path '%s'\n", targetPath)
			return
		}

		if err := clipboard.WriteAll(entry.Password); err != nil {
			shell.Printf("could not write password to clipboard: %s\n", err)
			return
		}
		shell.Println("password copied!")
	}
}
