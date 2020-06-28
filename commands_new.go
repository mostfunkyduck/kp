package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func NewEntry(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		path := strings.Split(args[0], "/")
		location, err := traversePath(currentLocation, strings.Join(path[0:len(path)-1], "/"))
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		shell.ShowPrompt(false)
		entry, err := location.NewEntry()
		if err != nil {
			shell.Printf("error creating new entry: %s\n", err)
			return
		}
		err = promptForEntry(shell, entry)
		shell.ShowPrompt(true)
		if err != nil {
			shell.Printf("could not collect user input: %s\n", err)
			if err := location.RemoveEntry(entry); err != nil {
				shell.Printf("could not remove malformed entry from group: %s\n", err)
			}
			return
		}
		entry.CreationTime = time.Now()
		entry.LastModificationTime = time.Now()
		entry.LastAccessTime = time.Now()

		savePath := shell.Get("filePath").(string)
		if savePath == "" {
			shell.Println("Database has been updated")
			return
		}

		shell.Printf("Database has been updated, save?: [Y/n] :  ")
		line, err := shell.ReadLineErr()
		if err != nil {
			shell.Printf("could not read user input: %s\n", err)
			return
		}

		if !strings.HasPrefix(line, "n") {
			db := shell.Get("db").(*keepass.Database)
			if err := saveDB(db, savePath); err != nil {
				shell.Printf("could not save database: %s\n", err)
				return
			}
		}
	}
}

func promptForEntry(shell *ishell.Shell, e *keepass.Entry) error {
	shell.Printf("Title: ")
	title, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("failed to read input: %s", err)
	}
	e.Title = title

	shell.Printf("URL: ")
	url, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("failed to read input: %s", err)
	}
	e.URL = url

	shell.Printf("username: ")
	un, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("failed to read input: %s", err)
	}
	e.Username = un

	var pw, pwConfirm string
	for {
		var err error
		shell.Printf("password: ")
		pw, err = shell.ReadPasswordErr()
		if err != nil {
			return fmt.Errorf("failed to read input: %s", err)
		}

		shell.Printf("enter password again: ")
		pwConfirm, err = shell.ReadPasswordErr()
		if err != nil {
			return fmt.Errorf("failed to read input: %s", err)
		}

		if pwConfirm != pw {
			shell.Println("password mismatch!")
			continue
		}
		break
	}

	e.Password = pw

	shell.Printf("Enter notes ('...' to terminate)\n\n")
	e.Notes = shell.ReadMultiLines("...")
	return nil
}
