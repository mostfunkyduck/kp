package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func listAttachment(entry t.Entry) (s string, err error) {
	attachment := entry.Get("attachment")
	if len(attachment.Value()) == 0 && attachment.Name() == "" {
		err = fmt.Errorf("entry has no attachment")
		return
	}
	s = fmt.Sprintf("Name: %s\nSize: %d bytes", attachment.Name(), len(attachment.Value()))
	return
}

func getAttachment(entry t.Entry, outputLocation string) (s string, err error) {
	f, err := os.Create(outputLocation)
	if err != nil {
		err = fmt.Errorf("could not open [%s]", outputLocation)
		return
	}
	defer f.Close()

	attachment := entry.Get("attachment")
	if len(attachment.Value()) == 0 {
		err = fmt.Errorf("entry has no attachment")
		return
	}
	written, err := f.Write(attachment.Value())
	if err != nil {
		err = fmt.Errorf("could not write to [%s]", outputLocation)
		return
	}

	s = fmt.Sprintf("wrote %s (%d bytes) to %s\n", attachment.Name(), written, outputLocation)
	return
}

func Attach(shell *ishell.Shell, cmd string) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		args := c.Args
		path := args[0]
		db := shell.Get("db").(t.Database)
		currentLocation := db.CurrentLocation()
		location, _, err := TraversePath(db, currentLocation, path)
		if err != nil {
			shell.Printf("error traversing path: %s\n", err)
			return
		}

		pieces := strings.Split(path, "/")
		name := pieces[len(pieces)-1]
		var intVersion int
		intVersion, err = strconv.Atoi(name)
		if err != nil {
			intVersion = -1 // assuming that this will never be a valid entry
		}
		for i, entry := range location.Entries() {

			if entry.Title() == name || (intVersion >= 0 && i == intVersion) {
				output, err := runAttachCommands(args, cmd, entry, shell)
				if err != nil {
					shell.Printf("could not run command [%s]: %s\n", cmd, err)
					return
				}
				shell.Println(output)
				return
			}
		}
		shell.Printf("could not find entry at path %s\n", path)
	}
}

func createAttachment(entry t.Entry, name string, path string) (output string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not open %s: %s", path, err)
	}

	entry.Set(c.NewValue(
		data,
		"attachment",
		false, false, false,
		t.BINARY,
	))

	return "added attachment to database", nil
}

// helper function run running attach commands. 'args' are all arguments after the attach command
// for instance, 'attach get foo bar' will result in args being '[foo, bar]'
func runAttachCommands(args []string, cmd string, entry t.Entry, shell *ishell.Shell) (output string, err error) {
	switch cmd {
	// attach create attachmentName /path/to/file
	case "create":
		if len(args) < 3 {
			return "", fmt.Errorf("bad syntax")
		}
		return createAttachment(entry, args[1], args[2])
	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("bad syntax")
		}

		outputLocation := args[1]
		if !confirmOverwrite(shell, outputLocation) {
			return "aborting", nil
		}
		return getAttachment(entry, outputLocation)
	case "details":
		return listAttachment(entry)
	default:
		return "", fmt.Errorf("invalid attach command")
	}
}
