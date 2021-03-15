package commands

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/atotto/clipboard"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"github.com/sethvargo/go-password/password"
)

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

// getEntryByPath returns the entry at path 'path' using context variables in shell 'shell'
func getEntryByPath(shell *ishell.Shell, path string) (entry t.Entry, ok bool) {
	db := shell.Get("db").(t.Database)
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
			// this is an edge case though, only happens if the UUID string is corrupted
			return nil, false
		}
		if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i ||
			entryName == string(entry.Title()) ||
			entryName == uuidString {
			return entry, true
		}
	}
	return nil, false
}

func isPresent(shell *ishell.Shell, path string) (ok bool) {
	db := shell.Get("db").(t.Database)
	l, e, err := TraversePath(db, db.CurrentLocation(), path)
	return err == nil && (l != nil || e != nil)
}

// doPrompt takes a t.Value, prompts for a new value, returns the value entered
func doPrompt(shell *ishell.Shell, value t.Value) (string, error) {
	var err error
	var input string
	switch value.Type() {
	case t.STRING:
		shell.Printf("%s: [%s]  ", value.Name(), value.FormattedValue(false))
		if value.Protected() {
			input, err = GetProtected(shell, string(value.Value()))
		} else {
			input, err = shell.ReadLineErr()
		}
	case t.BINARY:
		return "", fmt.Errorf("tried to edit binary directly")
	case t.LONGSTRING:
		shell.Printf("'%s' is a long text field, open in editor? [y/N] ", value.Name())
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
		return string(value.Value()), nil
	}

	return input, nil
}

// promptForEntry loops through all values in an entry, prompts to edit them, then applies any changes
func promptForEntry(shell *ishell.Shell, e t.Entry, title string) error {
	// make a copy of the entry's values for modification
	vals, err := e.Values()
	if err != nil {
		return fmt.Errorf("error retrieving values for entry '%s': %s", e.Title(), err)
	}
	valsToUpdate := []t.Value{}
	for _, value := range vals {
		if value != nil && !value.ReadOnly() && value.Type() != t.BINARY {
			newValue, err := doPrompt(shell, value)
			if err != nil {
				return fmt.Errorf("could not get value for %s, %s", value.Name(), err)
			}
			updatedValue := c.NewValue(
				[]byte(newValue),
				value.Name(),
				value.Searchable(),
				value.Protected(),
				value.ReadOnly(),
				value.Type(),
			)
			valsToUpdate = append(valsToUpdate, updatedValue)
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

		if err := PromptAndSave(shell); err != nil {
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

func GetLongString(value t.Value) (text string, err error) {
	// https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return "", err
	}

	filename := file.Name()

	defer os.Remove(filename)

	// start with what's already there
	if _, err = file.Write(value.Value()); err != nil {
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
		shell.Printf("password: ('g' to generate new)  ")
		pw, err = shell.ReadPasswordErr()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %s", err)
		}

		// default to whatever password was already set for the entry, if there is one
		if pw == "" && defaultPassword != "" {
			return defaultPassword, nil
		}

		// otherwise, we're either generating a new password or reading one from user input
		if pw == "g" {
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
func PromptAndSave(shell *ishell.Shell) error {

	shell.Printf("save database?: [Y/n]  ")
	line, err := shell.ReadLineErr()
	if err != nil {
		return fmt.Errorf("could not read user input: %s", err)
	}

	if line == "n" {
		shell.Println("continuing without saving")
		return nil
	}

	db := shell.Get("db").(t.Database)
	if err := db.Save(); err != nil {
		return fmt.Errorf("could not save database: %s", err)
	}

	// FIXME this should be a property of the DB, not a global

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
	switch strings.ToLower(entryData) {
	// FIXME hardcoded values
	case "username":
		// FIXME rewire this so that the entry provides the copy function
		data = string(entry.Get("username").Value())
	case "password":
		data = entry.Password()
	case "url":
		data = string(entry.Get("URL").Value())
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

// TraversePath, given a starting location and a UNIX-style path, will walk the path and return the final location or an error
// if the path points to an entry, the parent group is returned as well as the entry.
// If the path points to a group, the entry will be nil
func TraversePath(d t.Database, startingLocation t.Group, fullPath string) (finalLocation t.Group, finalEntry t.Entry, err error) {
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
