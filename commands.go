package main

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/atotto/clipboard"
	k "github.com/mostfunkyduck/kp/keepass"
	"github.com/sethvargo/go-password/password"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

// TODO break this function down, it's too long and mildly complicated
func openDB(shell *ishell.Shell) (db *keepass.Database, ok bool) {
	for {
		dbPath := *dbFile
		if envDbfile, found := os.LookupEnv("KP_DATABASE"); found && *dbFile == "" {
			dbPath = envDbfile
		}

		lockfilePath := fmt.Sprintf("%s.lock", dbPath)
		if _, err := os.Stat(lockfilePath); err == nil {
			shell.Printf("Lockfile exists for DB at path '%s', another process is using this database!\n", *dbFile)
			shell.Printf("Open anyways? Data loss may occur. (will only proceed if 'yes' is entered)  ")
			line, err := shell.ReadLineErr()
			if err != nil {
				shell.Printf("could not read user input: %s\n", line)
				return nil, false
			}

			if line != "yes" {
				shell.Println("aborting")
				return nil, false
			}
		}

		dbReader, err := os.Open(dbPath)
		if err != nil {
			shell.Printf("could not open db file [%s]: %s\n", dbPath, err)
			return nil, false
		}

		keyPath := *keyFile
		if envKeyfile, found := os.LookupEnv("KP_KEYFILE"); found && *keyFile == "" {
			keyPath = envKeyfile
		}

		var keyReader io.Reader
		if keyPath != "" {
			var err error
			keyReader, err = os.Open(keyPath)
			if err != nil {
				shell.Printf("could not open key file [%s]: %s\n", keyPath, err)
			}
		}

		password, passwordInEnv := os.LookupEnv("KP_PASSWORD")
		if !passwordInEnv {
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
			// if the password is coming from an environment variable, we need to terminate
			// after the first attempt or it will fall into an infinite loop
			if !passwordInEnv {
				// we are prompting for the password
				// in case this is a bad password, try again
				continue
			}
			return nil, false
		}

		if err := setLockfile(dbPath); err != nil {
			shell.Printf("could not create lock file at '%s': %s\n", lockfilePath, err)
			return nil, false
		}
		return db, true
	}
}

func setLockfile(filePath string) error {
	if filePath != "" {
		if _, err := os.Create(filePath + ".lock"); err != nil {
			return fmt.Errorf("could not create lock file at path '%s.lock': %s", filePath, err)
		}
	}
	return nil
}

// removeLockfile removes the lock file on the current savepath of the database
func removeLockfile(filePath string) error {
	if filePath != "" {
		if err := os.Remove(filePath + ".lock"); err != nil {
			return fmt.Errorf("could not remove lockfile: %s", err)
		}
	}
	return nil
}

// getEntryByPath returns the entry at path 'path' using context variables in shell 'shell'
func getEntryByPath(shell *ishell.Shell, path string) (entry k.Entry, ok bool) {
	db := shell.Get("db").(k.Database)
	location, entry, err := TraversePath(db, db.CurrentLocation(), path)
	if err != nil {
		return nil, false
	}

	if entry == nil {
		return nil, false
	}

	// a little extra work so that we can search by criteria other than 'title'
	// get the base name of the entry so that we can compare it to the actual
	// entries in this group
	entryNameBits := strings.Split(path, "/")
	entryName := entryNameBits[len(entryNameBits)-1]
	// loop so that we can compare entry indices
	for i, entry := range location.Entries() {
		uuidString, err := entry.UUIDString()
		if err != nil {
			// TODO we're swallowing this error :(
			// this is an edge case though
			return nil, false
		}
		if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i ||
			entryName == entry.Get("title").Value.(string) ||
			entryName == uuidString {
			return entry, true
		}
	}
	return nil, false
}

func isPresent(shell *ishell.Shell, path string) (ok bool) {
	db := shell.Get("db").(k.Database)
	l, e, err := TraversePath(db, db.CurrentLocation(), path)
	return err == nil && (l != nil || e != nil)
}

// doPrompt is a convenience function that takes a prompt and a default and
// then sets that up as an actual prompt for the user.
func doPrompt(shell *ishell.Shell, prompt string, deflt string) (string, error) {
	shell.Printf("%s: [%s]  ", prompt, deflt)
	input, err := shell.ReadLineErr()
	if err != nil {
		return "", fmt.Errorf("could not read user input: %s", err)
	}

	if input == "" {
		return deflt, nil
	}

	return input, nil
}

// TODO this will need to be refactored when we do v2
func promptForEntry(shell *ishell.Shell, e k.Entry, title string) error {
	var err error
	var url, un, pw, notes string
	// store all the changes in a temporary entry, don't update the target until all user input is collected
	if title, err = doPrompt(shell, "Title", title); err != nil {
		return fmt.Errorf("could not set title: %s", err)
	}

	if url, err = doPrompt(shell, "URL", e.Get("url").Value.(string)); err != nil {
		return fmt.Errorf("could not set URL: %s", err)
	}

	if un, err = doPrompt(shell, "Username", e.Get("username").Value.(string)); err != nil {
		return fmt.Errorf("could not set username: %s", err)
	}

	if pw, err = getPassword(shell, e.Get("password").Value.(string)); err != nil {
		return fmt.Errorf("could not set password: %s", err)
	}

	if notes, err = getNotes(shell, e.Get("notes").Value.(string)); err != nil {
		return fmt.Errorf("could not get notes: %s", err)
	}

	updated := e.Set(k.Value{Name: "title", Value: title})
	updated = e.Set(k.Value{Name: "url", Value: url}) || updated
	updated = e.Set(k.Value{Name: "username", Value: un}) || updated
	updated = e.Set(k.Value{Name: "password", Value: pw}) || updated
	updated = e.Set(k.Value{Name: "notes", Value: notes}) || updated

	if updated {
		shell.Println("edit successful, database has changed!")
		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("could not save: %s", err)
		}
	}
	return nil
}

// TODO this should be converted into a generic function for handling long-form value entry
// FIXME this code is also sucky and awkward
func getNotes(shell *ishell.Shell, existingNotes string) (string, error) {
	shell.Printf("Enter notes ('ctrl-c' to terminate)\nExisting notes:\n---\n%s\n---\n", existingNotes)
	// allow a single newline at the beginning to short circuit editing
	firstLine, err := shell.ReadLineErr()
	if err != nil {
		return existingNotes, fmt.Errorf("error reading user input: %s\n", err)
	}

	if firstLine == "" {
		return existingNotes, nil
	}

	// TODO kind of a hack - it doesn't seem to be able to match newlines, at least not in the WSL
	// TODO find a way to make this only use ctrl-c to terminate without funkiness
	newNotes := firstLine + "\n" + shell.ReadMultiLines("\n")

	if newNotes != existingNotes {
		shell.Println("Notes contents have changed, (o)verwrite, (A)ppend, or (d)iscard?\nOther edits will still be saved.")
		input, err := shell.ReadLineErr()
		if err != nil {
			return "", fmt.Errorf("could not read input on notes changes: %s", err)
		}

		input = strings.ToLower(input)
		switch input {
		case "d":
			shell.Println("discarding notes changes, other edits will be saved")
			return existingNotes, nil
		case "o":
			shell.Println("overwriting existing notes")
			return newNotes, nil
		default:
			shell.Println("appending to existing notes")
			return existingNotes + "\n" + newNotes, nil
		}
	}
	return newNotes, nil
}

func getPassword(shell *ishell.Shell, defaultPassword string) (pw string, err error) {
	for {
		shell.Printf("password: ('g' for automatic generation)  ")
		pw, err = shell.ReadPasswordErr()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %s", err)
		}

		// default to whatever password was already set for the entry
		if pw == "" {
			return defaultPassword, nil
		}

		// otherwise, we're either generating a new password or reading one from user input
		if pw == "g" {
			pw, err = password.Generate(20, 5, 5, false, false)
			if err != nil {
				return "", fmt.Errorf("failed to generate password: %s\n", err)
			}
			break
		}

		// the user is passing us a password, confirm it before saving
		shell.Printf("enter password again: ")
		pwConfirm, err := shell.ReadPasswordErr()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %s", err)
		}

		if pwConfirm != pw {
			shell.Println("password mismatch!")
			continue
		}
		break
	}
	return pw, nil
}

// promptAndSave prompts the user to save and returns whether or not they agreed to do so.
// it also makes sure that there's actually a path to save to
func promptAndSave(shell *ishell.Shell) error {

	shell.Printf("save database?: [Y/n]  ")
	line, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("could not read user input: %s", err)
	}

	if line == "n" {
		shell.Println("continuing without saving")
		return nil
	}

	db := shell.Get("db").(k.Database)
	if err := db.Save(); err != nil {
		return fmt.Errorf("could not save database: %s", err)
	}

	// FIXME this should be a property of the DB, not a global
	DBChanged = false
	shell.Println("database saved!")
	return nil
}

// copyFromEntry will find an entry and copy a given field in the entry
// to the clipboard
func copyFromEntry(shell *ishell.Shell, targetPath string, entryData string) error {
	entry, ok := getEntryByPath(shell, targetPath)
	if !ok {
		return fmt.Errorf("could not retrieve entry at path '%s'\n", targetPath)
	}

	var data string
	switch entryData {
	// FIXME hardcoded values
	case "username":
		data = entry.Get("username").Value.(string)
	case "password":
		data = entry.Get("password").Value.(string)
	case "url":
		data = entry.Get("url").Value.(string)
	default:
		return fmt.Errorf("'%s' was not a valid entry data type", entryData)
	}

	if data == "" {
		shell.Printf("warning! '%s' is an empty field!\n", entryData)
	}

	if err := clipboard.WriteAll(data); err != nil {
		return fmt.Errorf("could not write %s to clipboard: %s\n", entryData, err)
	}
	entry.SetLastAccessTime(time.Now())
	shell.Printf("%s copied!\n", entryData)
	shell.Println("(access time has been updated, will be persisted on next save)")
	return nil
}

// confirmOverwrite prompts the user about overwriting a given file
// it returns whether or not the user wants to overwrite
func confirmOverwrite(shell *ishell.Shell, path string) bool {
	shell.Printf("'%s' exists, overwrite? [y/N]  ", path)
	line, err := shell.ReadLineErr()
	if err != nil {
		shell.Printf("could not read user input: %s\n", line)
		return false
	}

	if line == "y" {
		shell.Println("overwriting")
		return true
	}
	return false
}

// TraversePath will, given a starting location and a UNIX-style path, will walk the path and return the final location or an error
// if the path points to an entry, the parent group is returned as well as the entry.
// If the path points to a group, the entry will be nil
func TraversePath(d k.Database, startingLocation k.Group, fullPath string) (finalLocation k.Group, finalEntry k.Entry, err error) {
	currentLocation := startingLocation
	root := d.Root()
	if fullPath == "/" {
		// short circuit now
		return root, nil, nil
	}

	if strings.HasPrefix(fullPath, "/") {
		// the user entered a fully qualified path, so start at the top
		currentLocation = root
	}

	// break the path up into components, remove terminal slashes since they don't actually do anything
	path := strings.Split(strings.TrimSuffix(fullPath, "/"), "/")
	// tracks whether or not the traversal encountered an entry
loop:
	for i, part := range path {
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
			return nil, nil, fmt.Errorf("tried to go to parent directory of '/'")
		}

		// regular traversal
		for _, group := range currentLocation.Groups() {
			// was the entity being searched for this group?
			if group.Name() == part {
				currentLocation = group
				continue loop
			}
		}

		for j, entry := range currentLocation.Entries() {
			// is the entity we're looking for this index or this entry?
			if entry.Get("title").Value.(string) == part || strconv.Itoa(j) == part {
				if i != len(path)-1 {
					// we encountered an entry before the end of the path, entries have no subgroups,
					// so this path is invalid
					return nil, nil, fmt.Errorf("invalid path '%s': '%s' is an entry, not a group", entry.Path(), fullPath)
				}
				// this is the end of the path, return the parent group and the entry
				return currentLocation, entry, nil
			}
		}
		// getting here means that we found neither a group nor an entry that matched 'part'
		// both of the loops looking for those short circuit when they find what they need
		return nil, nil, fmt.Errorf("could not find a group or entry named '%s'", part)
	}
	// we went all the way through the path and it points to currentLocation,
	// if it pointed to an entry, it would have returned above
	return currentLocation, nil, nil
}

