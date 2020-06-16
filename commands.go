package main

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"io"
	"os"
	"strconv"
	"strings"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		currentLocation := c.Get("currentLocation").(*keepass.Group)
		if len(args) == 0 {
			currentLocation = getRoot(currentLocation)
		} else {
			newLocation, err := traversePath(currentLocation, args[0])
			if err != nil {
				c.Err(fmt.Errorf("invalid path: %s", err))
				return
			}
			currentLocation = newLocation
		}
		shell.Set("currentLocation", currentLocation)
		c.SetPrompt(fmt.Sprintf("%s > ", currentLocation.Name))
	}
}

func Ls(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location := currentLocation
		if len(c.Args) > 0 {
			newLocation, err := traversePath(currentLocation, c.Args[0])
			if err != nil {
				c.Err(fmt.Errorf("Invalid path: %s", err))
				return
			}
			location = newLocation
		}

		lines := []string{}
		for _, group := range location.Groups() {
			lines = append(lines, fmt.Sprintf("%s/", group.Name))
		}
		for i, entry := range location.Entries() {
			lines = append(lines, fmt.Sprintf("%d: %s", i, entry.Title))
		}
		c.Println(strings.Join(lines, "\n"))
	}
}

func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			c.Err(fmt.Errorf("incorrect number of arguments to show"))
			return
		}

		fullMode := false
		entryName := c.Args[0]
		for _, arg := range c.Args {
			if strings.HasPrefix(arg, "-") {
				if arg == "-f" {
					fullMode = true
				}
				continue
			}
			entryName = arg
		}

		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, entryName)
		if err != nil {
			c.Err(fmt.Errorf("could not find entry named [%s]", entryName))
			return
		}

		// get the base name of the entry so that we can compare it to the actual
		// entries in this group
		entryNameBits := strings.Split(entryName, "/")
		entryName = entryNameBits[len(entryNameBits)-1]
		if *debugMode {
			shell.Printf("looking for entry [%s]", entryName)
		}
		for i, entry := range location.Entries() {
			if *debugMode {
				shell.Printf("looking at entry/idx for entry %s/%d\n", entry.Title, i, entryName)
			}
			if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i {
				outputEntry(*entry, c, fullMode)
				break
			}

			if entryName == entry.Title {
				outputEntry(*entry, c, fullMode)
				break
			}
		}
	}
}

func outputEntry(e keepass.Entry, c *ishell.Context, full bool) {
	c.Println(fmt.Sprintf("Title: %s", e.Title))
	c.Println(fmt.Sprintf("URL: %s", e.URL))
	c.Println(fmt.Sprintf("Username: %s", e.URL))
	password := "[redacted]"
	if full {
		password = e.Password
	}
	c.Println(fmt.Sprintf("Password: %s", password))
	c.Println(fmt.Sprintf("Notes : %s", e.Notes))
	if e.HasAttachment() {
		c.Println(fmt.Sprintf("Attachment: %s", e.Attachment.Name))
	}

}

func getRoot(location *keepass.Group) (root *keepass.Group) {
	for c := location; c.Parent() != nil; c = c.Parent() {
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

		shell.Print("enter database password: ")
		password, err := shell.ReadPasswordErr()
		if err != nil {
			shell.Printf("could not obtain password: %s\n", err)
			return nil, false
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
