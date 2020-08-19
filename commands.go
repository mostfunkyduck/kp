package main

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/atotto/clipboard"
	k "github.com/mostfunkyduck/kp/keepass"
	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	v2 "github.com/mostfunkyduck/kp/keepass/keepassv2"
	"github.com/sethvargo/go-password/password"
	keepass2 "github.com/tobischo/gokeepasslib/v3"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

// TODO break this function down, it's too long and mildly complicated
func openV2DB(shell *ishell.Shell) (db k.Database, ok bool) {
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

		creds, err := getV2Credentials(shell)
		if err != nil {
			shell.Println(err.Error())
			return nil, false
		}

		db := keepass2.NewDatabase(
			keepass2.WithDatabaseKDBXVersion4(),
		)
		db.Credentials = creds
		err = keepass2.NewDecoder(dbReader).Decode(db)
		if err != nil {
			shell.Printf("could not open database: %s\n", err)
			// if the password is coming from an environment variable, we need to terminate
			// after the first attempt or it will fall into an infinite loop
			_, passwordInEnv := os.LookupEnv("KP_PASSWORD")
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
		return v2.NewDatabase(db, shell.Get("filePath").(string), k.Options{}), true
	}
}

func openDB(shell *ishell.Shell) (db k.Database, ok bool) {
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

		// FIXME we want this to use v2 unless v1 is specified
		// FIXME we also want to decompose this function

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
		return v1.NewDatabase(db, shell.Get("filePath").(string)), true
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
			entryName == string(entry.Get("title").Value) ||
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

// doPrompt takes a k.Value, prompts for a new value, returns the value entered
func doPrompt(shell *ishell.Shell, value k.Value) (string, error) {
	var err error
	var input string
	switch value.Type {
	case k.STRING:
		shell.Printf("%s: [%s]  ", value.Name, string(value.Value))
		if value.Protected {
			input, err = GetProtected(shell, string(value.Value))
		} else {
			input, err = shell.ReadLineErr()
		}
	case k.BINARY:
		return "", fmt.Errorf("tried to edit binary directly")
	case k.LONGSTRING:
		shell.Printf("'%s' is a long text field, open in editor? [y/N]\n", value.Name)
		edit, err1 := shell.ReadLineErr()
		if err1 != nil {
			return "", fmt.Errorf("could not read user input: %s", err)
		}
		if edit == "y" {
			input, err = GetLongString(value)
			// normally, the user will see their input echoed, but not if an editor was open
			shell.Println(input)
		}
	}
	if err != nil {
		return "", fmt.Errorf("could not read user input: %s", err)
	}

	if input == "" {
		return string(value.Value), nil
	}

	return input, nil
}

// promptForEntry loops through all values in an entry, prompts to edit them, then applies any changes
func promptForEntry(shell *ishell.Shell, e k.Entry, title string) error {
	// make a copy of the entry's values for modification
	vals := e.Values()
	valsToUpdate := []k.Value{}
	for _, value := range vals {

		if !value.ReadOnly && !(value.Type == k.BINARY) {
			newValue, err := doPrompt(shell, value)
			if err != nil {
				return fmt.Errorf("could not get value for %s, %s", value.Name, err)
			}
			value.Value = []byte(newValue)
			valsToUpdate = append(valsToUpdate, value)
		}
	}

	// determine whether any of the provided values was an actual update meriting a save
	updated := false
	for _, value := range valsToUpdate {
		if e.Set(value) {
			updated = true
		}
	}

	if updated {
		shell.Println("edit successful, database has changed!")
		DBChanged = true
		if err := promptAndSave(shell); err != nil {
			shell.Printf("could not save: %s", err)
		}
	}
	return nil
}

// OpenFileInEditor opens filename in a text editor.
func OpenFileInEditor(filename string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // because use vim or you're a troglodyte
	}

	// Get the full executable path for the editor.
	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
func GetLongString(value k.Value) (text string, err error) {
	// https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return "", err
	}

	filename := file.Name()

	defer os.Remove(filename)

	// start with what's already there
	if _, err = file.Write(value.Value); err != nil {
		return "", err
	}

	if err = file.Close(); err != nil {
		return "", err
	}

	if err = OpenFileInEditor(filename); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func GetProtected(shell *ishell.Shell, defaultPassword string) (pw string, err error) {
	for {
		shell.Printf("password: ('g' to generate, empty passwords will automatically be regenerated)  ")
		pw, err = shell.ReadPasswordErr()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %s", err)
		}

		// default to whatever password was already set for the entry, if there is one
		if pw == "" && defaultPassword != "" {
			return defaultPassword, nil
		}


		// otherwise, we're either generating a new password or reading one from user input
		if pw == "g" || pw == "" {
			// FIXME (low pri for now) needs better generation than hardcoding the number of syms
			pw, err = password.Generate(20, 5, 5, false, false)
			if err != nil {
				return "", fmt.Errorf("failed to generate password: %s\n", err)
			}
			shell.Println("generated new password")
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
		data = string(entry.Get("username").Value)
	case "password":
		data = string(entry.Get("password").Value)
	case "url":
		data = string(entry.Get("url").Value)
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
			if string(entry.Title()) == part || strconv.Itoa(j) == part {
				if i != len(path)-1 {
					// we encountered an entry before the end of the path, entries have no subgroups,
					// so this path is invalid

					return nil, nil, fmt.Errorf("invalid path '%s': '%s' is an entry, not a group", entry.Title(), fullPath)
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

// getV2Credentials builds a keepass2 db credentials object based on the cli arguments
// and environment variables
func getV2Credentials(shell *ishell.Shell) (*keepass2.DBCredentials, error) {
	password, passwordInEnv := os.LookupEnv("KP_PASSWORD")
	if !passwordInEnv {
		shell.Print("enter database password: ")
		var err error
		password, err = shell.ReadPasswordErr()
		if err != nil {
			return &keepass2.DBCredentials{}, fmt.Errorf("could not obtain password: %s\n", err)
		}
	}

	creds := keepass2.NewPasswordCredentials(password)

	keyPath := *keyFile // this is the flag in main.go
	if envKeyfile, found := os.LookupEnv("KP_KEYFILE"); found && *keyFile == "" {
		keyPath = envKeyfile
	}

	if keyPath != "" {
		var err error
		creds.Key, err = keepass2.ParseKeyFile(keyPath)
		if err != nil {
			return &keepass2.DBCredentials{}, fmt.Errorf("could not parse key file [%s]: %s\n", keyPath, err)
		}
	}
	return creds, nil
}
