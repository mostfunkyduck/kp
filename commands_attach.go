package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func listAttachment(entry *keepass.Entry) (s string, err error) {
	s = fmt.Sprintf("Name: %s\nSize: %d bytes", entry.Attachment.Name, len(entry.Attachment.Data))
	return
}

func getAttachment(entry *keepass.Entry, outputLocation string) (s string, err error) {
	f, err := os.Create(outputLocation)
	if err != nil {
		return "", fmt.Errorf("could not open [%s]", outputLocation)
	}
	defer f.Close()

	written, err := f.Write(entry.Attachment.Data)
	if err != nil {
		return "", fmt.Errorf("could not write to [%s]", outputLocation)
	}

	s = fmt.Sprintf("wrote %s (%d bytes) to %s\n", entry.Attachment.Name, written, outputLocation)
	return s, nil
}

func Attach(shell *ishell.Shell, cmd string) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		args := c.Args
		path := args[0]
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, path)
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

			if entry.Title == name || (intVersion >= 0 && i == intVersion) {
				output, err := runAttachCommands(args, cmd, entry)
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

// helper function run running attach commands. 'args' are all arguments after the attach command
// for instance, 'attach get foo bar' will result in args being '[foo, bar]'
func runAttachCommands(args []string, cmd string, entry *keepass.Entry) (output string, err error) {
	switch cmd {
	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("bad syntax")
		}

		return getAttachment(entry, args[1])
	case "details":
		return listAttachment(entry)
	default:
		return "", fmt.Errorf("invalid attach command")
	}
}
