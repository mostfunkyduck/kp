package main

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"io"
	"os"
	"strconv"
	"strings"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

const (
	INCORRECT_NUMBER_OF_ARGUMENTS int = iota
	INVALID_ATTACH_COMMAND
	INVALID_ARGUMENTS
	INVALID_ARGUMENTS_ATTACH_GET
)

// these will be base error messages, they can be spruced up with fmt.Sprintf()
var ERROR_MESSAGE = map[int]string{
	INCORRECT_NUMBER_OF_ARGUMENTS: "incorrect number of arguments",
	INVALID_ARGUMENTS:             "invalid arguments",
	INVALID_ATTACH_COMMAND:        "invalid attach command",
	INVALID_ARGUMENTS_ATTACH_GET:  "syntax: attach get <entry> <filesystem location>",
	//INVALID_PATH:	"invalid path",
}

func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		if len(args) == 0 {
			currentLocation = getRoot(currentLocation)
		} else {
			newLocation, err := traversePath(currentLocation, args[0])
			if err != nil {
				shell.Println(fmt.Sprintf("invalid path: %s", err))
				return
			}
			currentLocation = newLocation
		}
		shell.Set("currentLocation", currentLocation)
		shell.SetPrompt(fmt.Sprintf("%s > ", currentLocation.Name))
	}
}

func Ls(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location := currentLocation
		entityName := "/"
		if len(c.Args) > 0 {
			path := strings.Split(c.Args[0], "/")
			entityName = path[len(path)-1]
			newLocation, err := traversePath(currentLocation, c.Args[0])
			if err != nil {
				shell.Printf("Invalid path: %s", err)
				return
			}
			location = newLocation
		}

		lines := []string{}
		for _, group := range location.Groups() {
			if group.Name == entityName {
				shell.Println(group.Name + "/")
				return
			}
			lines = append(lines, fmt.Sprintf("%s/", group.Name))
		}
		for i, entry := range location.Entries() {
			entryLine := fmt.Sprintf("%d: %s", i, entry.Title)
			lines = append(lines, entryLine)
			if entry.Title == entityName {
				shell.Println(entryLine)
				return
			}
		}
		shell.Println(strings.Join(lines, "\n"))
	}
}

func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println(ERROR_MESSAGE[INCORRECT_NUMBER_OF_ARGUMENTS])
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

		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, path)
		if err != nil {
			shell.Println(fmt.Sprintf("could not find entry named [%s]", path))
			return
		}

		// get the base name of the entry so that we can compare it to the actual
		// entries in this group
		entryNameBits := strings.Split(path, "/")
		entryName := entryNameBits[len(entryNameBits)-1]
		if *debugMode {
			shell.Printf("looking for entry [%s]", entryName)
		}
		for i, entry := range location.Entries() {
			if *debugMode {
				shell.Printf("looking at entry/idx for entry %s/%d\n", entry.Title, i, entryName)
			}
			if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i {
				outputEntry(*entry, shell, path, fullMode)
				break
			}

			if entryName == entry.Title {
				outputEntry(*entry, shell, path, fullMode)
				break
			}
		}
	}
}

func listAttachment(entry *keepass.Entry) (s string, err error) {
	s = fmt.Sprintf("Name: %s, Size: %d bytes", entry.Attachment.Name, len(entry.Attachment.Data))
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
func Attach(shell *ishell.Shell) (f func(c *ishell.Context)) {
	attachCommands := map[string]bool{
		"get":    true,
		"list":   true,
		"cancel": true,
	}
	return func(c *ishell.Context) {
		if len(c.Args) < 2 {
			shell.Println(ERROR_MESSAGE[INVALID_ARGUMENTS])
			return
		}

		args := c.Args
		cmd := args[0]
		path := args[1]
		if _, ok := attachCommands[cmd]; !ok {
			shell.Printf("%s: %s\n", INVALID_ATTACH_COMMAND, cmd)
			return
		}
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
func outputEntry(e keepass.Entry, s *ishell.Shell, path string, full bool) {
	s.Println(fmt.Sprintf("Location: %s", path))
	s.Println(fmt.Sprintf("Title: %s", e.Title))
	s.Println(fmt.Sprintf("URL: %s", e.URL))
	s.Println(fmt.Sprintf("Username: %s", e.Username))
	password := "[redacted]"
	if full {
		password = e.Password
	}
	s.Println(fmt.Sprintf("Password: %s", password))
	s.Println(fmt.Sprintf("Notes: %s", e.Notes))
	if e.HasAttachment() {
		s.Println(fmt.Sprintf("Attachment: %s", e.Attachment.Name))
	}

}

func getRoot(location *keepass.Group) (root *keepass.Group) {
	for root = location; root.Parent() != nil; root = root.Parent() {
	}
	return root
}

// given a starting location and a UNIX-style path, will walk the path and return the final location or an error
func traversePath(startingLocation *keepass.Group, fullPath string) (finalLocation *keepass.Group, err error) {
	currentLocation := startingLocation
	root := getRoot(currentLocation)
	if fullPath == "/" {
		// short circuit now
		return root, nil
	}

	if strings.HasPrefix(fullPath, "/") {
		// the user entered a fully qualified path, so start at the top
		currentLocation = root
	}

	// break the path up into components
	path := strings.Split(fullPath, "/")
	for _, part := range path {
		if part == "." || part == "" {
			continue
		}

		if part == ".." {
			// if we're not at the root, go up a level
			if currentLocation.Parent() != nil {
				currentLocation = currentLocation.Parent()
				continue
			}
			// we're at the root, the user wanted to go higher, that's no bueno
			return nil, fmt.Errorf("root group has no parent")
		}
		// regular traversal
		found := false
		for _, group := range currentLocation.Groups() {
			// was the entity being searched for this group?
			if group.Name == part {
				currentLocation = group
				found = true
				break
			}
		}
		for i, entry := range currentLocation.Entries() {
			// is the entity we're looking for this index or this entry?
			if entry.Title == part || strconv.Itoa(i) == part {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("could not find a group or entry named [%s]", part)
		}
	}
	return currentLocation, nil
}

func openDB(shell *ishell.Shell) (db *keepass.Database, ok bool) {
	if *dbFile == "" {
		shell.Println("no db file provided!")
		return nil, false
	}

	for {
		dbReader, err := os.Open(*dbFile)
		if err != nil {
			shell.Print("could not open db file [%s]: %s\n", *dbFile, err)
			return nil, false
		}

		var keyReader io.Reader
		if *keyFile != "" {
			keyReader, err = os.Open(*keyFile)
			if err != nil {
				shell.Print("could not open key file %s: %s\n", *keyFile, err)
			}
		}

		password := os.Getenv("KP_PASSWORD")
		if password == "" {
			shell.Print("enter database password: ")
			var err error
			password, err = shell.ReadPasswordErr()
			if err != nil {
				shell.Printf("could not obtain password: %s\n", err)
				return nil, false
			}
		}

		if *debugMode {
			shell.Printf("got password: %s\n", password)
		}

		opts := &keepass.Options{
			Password: password,
			KeyFile:  keyReader,
		}

		db, err := keepass.Open(dbReader, opts)
		if err != nil {
			shell.Printf("could not open database: %s\n", err)
			// in case this is a bad password, try again
			continue
		}
		return db, true
	}

}

func runAttachCommands(args []string, cmd string, entry *keepass.Entry) (output string, err error) {
	switch cmd {
	case "get":
		if len(args) < 3 {
			return "", fmt.Errorf(ERROR_MESSAGE[INVALID_ARGUMENTS_ATTACH_GET])
		}
		return getAttachment(entry, args[2])
	case "list":
		return listAttachment(entry)
	default:
		return "", fmt.Errorf(ERROR_MESSAGE[INVALID_ATTACH_COMMAND])
	}
}
