package commands

import (
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func NewEntry(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		if isPresent(shell, args[0]) {
			shell.Printf("cannot create duplicate entity '%s'\n", args[0])
			return
		}

		db := shell.Get("db").(t.Database)

		pathBits := strings.Split(args[0], "/")
		parentPath := strings.Join(pathBits[0:len(pathBits)-1], "/")
		location, entry, err := TraversePath(db, db.CurrentLocation(), parentPath)
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		if entry != nil {
			shell.Printf("entry '%s' already exists!\n", entry.Title())
			return
		}

		if location.IsRoot() {
			shell.Println("cannot add entries to root node")
			return
		}

		shell.ShowPrompt(false)
		entry, err = location.NewEntry(pathBits[len(pathBits)-1])
		if err != nil {
			shell.Printf("error creating new entry: %s\n", err)
			return
		}
		entry.SetCreationTime(time.Now())
		entry.SetLastModificationTime(time.Now())
		entry.SetLastAccessTime(time.Now())

		err = promptForEntry(shell, entry, entry.Title())
		shell.ShowPrompt(true)
		if err != nil {
			shell.Printf("could not collect user input: %s\n", err)
			if err := location.RemoveEntry(entry); err != nil {
				shell.Printf("could not remove malformed entry from group: %s\n", err)
			}
			return
		}
	}
}
