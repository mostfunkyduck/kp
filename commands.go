package main

// All the commands that the shell will run
// Note: do NOT use context.Err() here, it will impede testing.

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func NewGroup(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		path := strings.Split(c.Args[0], "/")
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, strings.Join(path[0:len(path)-1], "/"))
		if err != nil {
			shell.Printf("invalid path: " + err.Error())
			return
		}
		location.NewSubgroup().Name = path[len(path)-1]
	}
}
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
		e, err := location.NewEntry()
		if err != nil {
			shell.Printf("could not create new entry: %s\n", err)
			return
		}
		e.Title = path[len(path)-1]
		e.LastModificationTime = time.Now()
		e.CreationTime = time.Now()
		e.LastAccessTime = time.Now()
	}
}
func Cd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		args := c.Args
		currentLocation := shell.Get("currentLocation").(*keepass.Group)
		if len(args) == 0 {
			// FIXME db.Root() can come in from the context
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

// FIXME the keepass library has a bug where you can't get the parent
// unless the entry is a pointer to the one in the db (it's comparing pointer values)
// this can/should/will be fixed in my fork
func searchEntries(g keepass.Group, term *regexp.Regexp) (titles []string) {
	for _, e := range g.Entries() {
		if term.FindString(e.Title) != "" ||
			term.FindString(e.Notes) != "" ||
			term.FindString(e.Attachment.Name) != "" ||
			term.FindString(e.Username) != "" {
			titles = append(titles, e.Title)
		}
	}
	return titles
}

// searchGroup returns a list of paths to entries or groups matching the search terms
func searchGroup(g keepass.Group, term *regexp.Regexp, path string) (paths []string) {
	// the initial group will send in "", meaning it should be skipped in the path output
	if path != "" {
		path = path + "/" + g.Name
		if term.FindString(g.Name) != "" {
			// adding a terminal / to indicate that this is a group (imitating how directories are output in ls by default
			paths = append(paths, path+"/")
		}
	} else {
		path = "."
	}

	for _, title := range searchEntries(g, term) {
		paths = append(paths, path+"/"+title)
	}
	for _, g := range g.Groups() {
		paths = append(paths, searchGroup(*g, term, path)...)
	}
	return paths
}

// This implements the equivalent of kpcli's "find" command, just with a name
// that won't be confused for the shell command of the same name
func Search(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := getRoot(shell.Get("currentLocation").(*keepass.Group))
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		term, err := regexp.Compile(c.Args[0])
		if err != nil {
			shell.Printf("could not compile search term into a regular expression: %s", err)
			return
		}

		// kpcli makes a fake group for search results, which gets into trouble when entries have the same name in different paths
		// this takes a different approach of printing out full paths and letting the user type them in later
		// a little more typing for the user, less oddness in the implementation though
		for _, result := range searchGroup(*currentLocation, term, "") {
			// the tab makes it a little more readable
			shell.Printf("\t%s\n", result)
		}
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

func generateRandomString(length int) (str string) {
	// based on a few things, mainly https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	rand.Seed(time.Now().UnixNano())
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-_=+\\][{}|/.,?><'"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func syntaxCheck(c *ishell.Context, numArgs int) (errorString string, ok bool) {
	if len(c.Args) < numArgs {
		return "syntax: " + c.Cmd.Help, false
	}
	return "", true
}

// loadOrGenerateKey will return a file handle for the key, will prompt the user to generate a key if they so desire.
// if the key doesn't exist and the user declines to generate it, will return a nil reader and a nil error
func loadOrGenerateKey(shell *ishell.Shell, path string) (f io.Reader, err error) {
	if _, err := os.Stat(path); err != nil {
		shell.Printf("%s does not exist: generate a key at that location? [yes]\n", path)
		shell.ShowPrompt(false)
		choice := shell.ReadLine()
		shell.ShowPrompt(true)
		if choice != "yes" {
			shell.Println("aborting operation")
			return nil, nil
		}
		str := generateRandomString(2048)
		if err := ioutil.WriteFile(path, []byte(str), 0644); err != nil {
			return nil, err
		}
	}

	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	return
}

func SaveAs(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}
		var file io.Reader
		if len(c.Args) >= 2 {
			_file, err := loadOrGenerateKey(shell, c.Args[1])
			if err != nil {
				shell.Printf("could not load or generate key: %s\n", err)
				return
			}
			// this will either set the reader in the outer scope or set it to nil
			// nil is fine, zero values won't hurt later
			file = _file
		}

		db := shell.Get("db").(*keepass.Database)
		opts := &keepass.Options{
			// FIXME prompt for this
			// Password: "yaakov is testing",
			KeyFile: file,
		}
		if err := db.SetOpts(opts); err != nil {
			shell.Printf("could not set DB options: %s", err)
			return
		}

		savePath := c.Args[0]
		w, err := os.Create(savePath)
		if err != nil {
			shell.Printf("could not open/create db save location [%s]: %s", savePath, err)
			return
		}
		if err = db.Write(w); err != nil {
			shell.Printf("error writing database to [%s]: %s", savePath, err)
			return
		}
	}
}
func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
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
			if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i ||
				entryName == entry.Title ||
				entryName == entry.UUID.String() {
				outputEntry(*entry, shell, path, fullMode)
				return
			}
		}
	}
}

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

func formatTime(t time.Time) (formatted string) {
	timeFormat := "Mon Jan 2 15:04:05 MST 2006"
	if (t == time.Time{}) {
		formatted = "unknown"
	} else {
		since := time.Since(t).Round(time.Duration(1) * time.Second)
		sinceString := since.String()

		// greater than or equal to 1 day
		if since.Hours() >= 24 {
			sinceString = fmt.Sprintf("%d days ago", int(since.Hours()/24))
		}

		// greater than or equal to ~1 month
		if since.Hours() >= 720 {
			// rough estimate, not accounting for non-30-day months
			months := int(since.Hours() / 720)
			sinceString = fmt.Sprintf("about %d months ago", months)
		}

		// greater or equal to 1 year
		if since.Hours() >= 8760 {
			// yes yes yes, leap years aren't 365 days long
			years := int(since.Hours() / 8760)
			sinceString = fmt.Sprintf("about %d years ago", years)
		}

		// less than a second
		if since.Seconds() < 1.0 {
			sinceString = "less than a second ago"
		}

		formatted = fmt.Sprintf("%s (%s)", t.Local().Format(timeFormat), sinceString)
	}
	return
}

func outputEntry(e keepass.Entry, s *ishell.Shell, path string, full bool) {
	s.Printf("UUID: %s\n", e.UUID)

	s.Printf("Creation Time: %s\n", formatTime(e.CreationTime))
	s.Printf("Last Modified: %s\n", formatTime(e.LastModificationTime))
	s.Printf("Last Accessed: %s\n", formatTime(e.LastAccessTime))
	s.Printf("Location: %s\n", path)
	s.Printf("Title: %s\n", e.Title)
	s.Printf("URL: %s\n", e.URL)
	s.Printf("Username: %s\n", e.Username)
	password := "[redacted]"
	if full {
		password = e.Password
	}
	s.Printf("Password: %s\n", password)
	s.Printf("Notes: %s\n", e.Notes)
	if e.HasAttachment() {
		s.Printf("Attachment: %s\n", e.Attachment.Name)
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
