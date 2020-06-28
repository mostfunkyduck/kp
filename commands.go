package main

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

func saveDB(db *keepass.Database, savePath string) error {
	w, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("could not open/create db save location [%s]: %s", savePath, err)
	}
	if err = db.Write(w); err != nil {
		return fmt.Errorf("error writing database to [%s]: %s", savePath, err)
	}
	return nil
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
	for {
		dbReader, err := os.Open(*dbFile)
		if err != nil {
			shell.Printf("could not open db file [%s]: %s\n", *dbFile, err)
			return nil, false
		}

		var keyReader io.Reader
		if *keyFile != "" {
			keyReader, err = os.Open(*keyFile)
			if err != nil {
				shell.Printf("could not open key file %s: %s\n", *keyFile, err)
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
