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
	"github.com/sethvargo/go-password/password"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

var backupExtension = ".kpbackup"

func backupDB(db *keepass.Database, savePath string) error {
	backupPath := savePath + backupExtension
	w, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("could not open file '%s': %s", backupPath, err)
	}

	if err := db.Write(w); err != nil {
		return fmt.Errorf("could not write to file '%s': %s", backupPath, err)
	}
	return nil
}

func removeBackup(savePath string) error {
	backupPath := savePath + backupExtension
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("could not remove backup file '%s': %s", backupPath, err)
	}
	return nil
}

func saveDB(db *keepass.Database, savePath string) error {
	if err := backupDB(db, savePath); err != nil {
		return fmt.Errorf("could not back up database: %s", err)
	}
	w, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("could not open/create db save location [%s]: %s", savePath, err)
	}

	if err = db.Write(w); err != nil {
		return fmt.Errorf("error writing database to [%s]: %s", savePath, err)
	}

	if err := removeBackup(savePath); err != nil {
		return fmt.Errorf("could not remove backup after saving: %s", err)
	}
	return nil
}

func getRoot(location *keepass.Group) (root *keepass.Group) {
	for root = location; root.Parent() != nil; root = root.Parent() {
	}
	return root
}

// traversePath will, given a starting location and a UNIX-style path, will walk the path and return the final location or an error
// if the path points to an entry, the parent group is returned, otherwise the target group is returned
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
	lockfilePath := fmt.Sprintf("%s.lock", *dbFile)
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
		envPassword := os.Getenv("KP_PASSWORD")
		password := envPassword
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
			if envPassword == "" {
				// we are prompting for the password
				// in case this is a bad password, try again
				continue
			}
			return nil, false
		}

		if err := setLockfile(shell); err != nil {
			shell.Printf("could not create lock file at '%s': %s\n", lockfilePath, err)
			return nil, false
		}
		return db, true
	}
}

func setLockfile(shell *ishell.Shell) error {
	filePath := shell.Get("filePath").(string)
	if filePath != "" {
		if _, err := os.Create(filePath + ".lock"); err != nil {
			return fmt.Errorf("could not create lock file at path '%s.lock': %s", filePath, err)
		}
	}
	return nil
}

func removeLockfile(shell *ishell.Shell) error {
	filePath := shell.Get("filePath").(string)
	if filePath != "" {
		if err := os.Remove(filePath + ".lock"); err != nil {
			return fmt.Errorf("could not remove lockfile: %s", err)
		}
	}
	return nil
}

// getEntryByPath returns the entry at path 'path' using context variables in shell 'shell'
func getEntryByPath(shell *ishell.Shell, path string) (entry *keepass.Entry, ok bool) {
	currentLocation := shell.Get("currentLocation").(*keepass.Group)
	location, err := traversePath(currentLocation, path)
	if err != nil {
		return nil, false
	}

	// get the base name of the entry so that we can compare it to the actual
	// entries in this group
	entryNameBits := strings.Split(path, "/")
	entryName := entryNameBits[len(entryNameBits)-1]
	for i, entry := range location.Entries() {
		if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i ||
			entryName == entry.Title ||
			entryName == entry.UUID.String() {
			return entry, true
		}
	}
	return nil, false
}

func isPresent(shell *ishell.Shell, path string) (ok bool) {
	currentLocation := shell.Get("currentLocation").(*keepass.Group)
	_, err := traversePath(currentLocation, path)
	return err == nil
}

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

func promptForEntry(shell *ishell.Shell, e *keepass.Entry, title string) error {
	var err error
	if e.Title, err = doPrompt(shell, "Title", title); err != nil {
		return fmt.Errorf("could not set title: %s", err)
	}

	if e.URL, err = doPrompt(shell, "URL", e.URL); err != nil {
		return fmt.Errorf("could not set URL: %s", err)
	}

	if e.Username, err = doPrompt(shell, "Username", e.Username); err != nil {
		return fmt.Errorf("could not set username: %s", err)
	}

	if e.Password, err = getPassword(shell, e.Password); err != nil {
		return fmt.Errorf("could not set password: %s", err)
	}

	shell.Printf("Enter notes ('ctrl-c' to terminate) ['%s']\n\n", e.Notes)
	// TODO kind of a hack - it doesn't seem to be able to match newlines, at least not in the WSL
	// TODO find a way to make this only use ctrl-c to terminate without funkiness
	newNotes := shell.ReadMultiLines("\n")

	if newNotes != e.Notes {
		shell.Printf("Notes contents have changed, overwrite existing? [Y/n]  ")
		input, err := shell.ReadLineErr()
		if err != nil {
			return fmt.Errorf("could not read input on notes changes: %s", err)
		}

		if input == "n" {
			shell.Println("\ndiscarding notes changes, other edits will be saved")
			return nil
		}
	}
	e.Notes = newNotes
	return nil
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

// getPwd will walk up the group hierarchy to determine the path to the current location
func getPwd(shell *ishell.Shell, group *keepass.Group) (fullPath string) {
	for ; group != nil; group = group.Parent() {
		if group.IsRoot() {
			fullPath = "/" + fullPath
			break
		}
		fullPath = group.Name + "/" + fullPath
	}
	return fullPath
}

// promptAndSave prompts the user to save and returns whether or not they agreed to do so.
// it also makes sure that there's actually a path to save to
func promptAndSave(shell *ishell.Shell) error {
	filePath := shell.Get("filePath").(string)
	if filePath == "" {
		return fmt.Errorf("no file path for database")
	}

	shell.Printf("save database?: [Y/n]  ")
	line, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("could not read user input: %s", err)
	}

	if line == "n" {
		shell.Println("continuing without saving")
		return nil
	}

	db := shell.Get("db").(*keepass.Database)
	if err := saveDB(db, filePath); err != nil {
		return fmt.Errorf("could not save database: %s", err)
	}
	shell.Println("database saved!")
	return nil
}
