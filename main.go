package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/abiosoft/ishell"
	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	v2 "github.com/mostfunkyduck/kp/internal/backend/keepassv2"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"github.com/mostfunkyduck/kp/internal/commands"
)

var (
	keyFile        = flag.String("key", "", "a key file to use to unlock the db")
	dbFile         = flag.String("db", "", "the db to open")
	keepassVersion = flag.Int("kpversion", 1, "which version of keepass to use (1 or 2)")
	version        = flag.Bool("version", false, "print version and exit")
	noninteractive = flag.String("n", "", "execute a given command and exit")
)

/*
	All examples assume a structure of:
		one two/
		  three four/
				otherstuff
			stuff
	I had to hack the completions part of abiosoft/ishell to take proposed completions instead of taking all potential completions
	The library had breaking on spaces, the following behavior was observed:
		1.	inputting "one two" would result in "one" being counted as a "priorWord" and "two" being the "wordToComplete"
		2.	it would then treat the returned values as *potential* matches, which had to match the prefix, which is the argument
				labeled "wordToComplete"
		3.	this meant that when "one two" came in, you COULD piece results back together as "one two/stuff", but returning that
				would not result in matches since it would only return things that started with "two"
		4.	This also meant that even if you stripped the part before the space (i.e. returning "two/stuff"), the results displayed would be
		  	prefixed with "two", which broke nested search
	Therefore, the hack on the ishell library takes the actual completions, so:
		1.	"one two/st<tab>" will return "uff"
		2.	"one" will return "one two/"
		3.  "one two/" will return "stuff" and "three four/"
*/
func fileCompleter(shell *ishell.Shell, printEntries bool) func(string, []string) []string {
	return func(wordToComplete string, priorWords []string) (ret []string) {
		searchLocation := ""
		if len(priorWords) > 0 {
			// join together all the previous tokens so that we compensate for spaces
			// priorWords will be ["one"], wordToComplete will be "two", we want "one two"
			wordToComplete = strings.Join(priorWords, " ") + " " + wordToComplete
		}

		if wordToComplete[len(wordToComplete)-1] == '/' {
			// if the phrase we're completing is slash-terminated, it's a group that we're trying
			// to enumerate the contents of
			// i.e. "one two/" and we want to get "stuff"
			searchLocation = wordToComplete
		} else if strings.Contains(wordToComplete, "/") {
			// the wordToComplete is at least partially a path
			// i.e "one two/three", in which case we wanted to match things under "one two/" starting with "three"
			// this will strip out everything after the last "/" to find the search path
			rxp := regexp.MustCompile(`(.+\/).+$`)
			searchLocation = rxp.ReplaceAllString(wordToComplete, "$1")
		}

		// get the current location to search for matches
		db := shell.Get("db").(t.Database)
		location, _, err := commands.TraversePath(db, db.CurrentLocation(), searchLocation)
		// if the there was an error, assume path doesn't exist
		if err != nil {
			return
		}

		// helper function to identify completions
		f := func(token string) string {
			// trim the wordToComplete down to the part after the directory name
			// we have the directory name in searchLocation, we want to search that directory for the prefix
			// alternatively, we can determine that we're enumerating a directory, in which case no matching will be done
			reg := regexp.MustCompile(`.*/`)
			prefix := reg.ReplaceAllString(wordToComplete, "")

			if prefix == "" {
				// the wordToComplete was an entire path, we're enumerating the contents of a directory
				// return the token as-is since we don't have to do any matching to figure out potential completions
				// i.e. if the user inputted "one two/<tab>", we want to return "stuff" and "three four"
				return token
			}

			if strings.HasPrefix(token, prefix) {
				// the wordToComplete contained a partial prefix of an item in a directory to match
				// strip the already-matched prefix
				// i.e. if the user passed in "one two/st<tab>", we will scan "stuff" and "three four" and return "stuff"
				return strings.TrimPrefix(token, prefix)
			}
			return ""
		}

		// Loop through all the groups and entries in this group and check for matches
		for _, g := range location.Groups() {
			completion := f(g.Name() + "/")
			if completion != "" {
				ret = append(ret, completion)
			}
		}

		// loop through entries iff the command needs us to
		if printEntries {
			for _, e := range location.Entries() {
				completion := f(e.Title())
				if completion != "" {
					ret = append(ret, completion)
				}
			}
		}

		return ret
	}
}

func buildVersionString() string {
	return fmt.Sprintf("%s.%s-%s.%s (built on %s from %s)", VersionRelease, VersionBuildDate, VersionBuildTZ, VersionBranch, VersionHostname, VersionRevision)
}

// promptForDBPassword will determine the password based on environment vars or, lacking those, a prompt to the user
func promptForDBPassword(shell *ishell.Shell) (string, error) {
	// we are prompting for the password
	shell.Print("enter database password: ")
	return shell.ReadPasswordErr()
}

// newDB will create or open a DB with the parameters specified.  `open` indicates whether the DB should be opened or not (vs created)
func newDB(dbPath string, password string, keyPath string, version int) (t.Database, error) {
	var dbWrapper t.Database
	switch version {
	case 2:
		dbWrapper = &v2.Database{}
	case 1:
		dbWrapper = &v1.Database{}
	default:
		return nil, fmt.Errorf("invalid version '%d'", version)
	}
	dbOpts := t.Options{
		DBPath:   dbPath,
		Password: password,
		KeyPath:  keyPath,
	}
	err := dbWrapper.Init(dbOpts)
	return dbWrapper, err
}

func main() {
	flag.Parse()

	shell := ishell.New()
	if *version {
		shell.Printf("version: %s\n", buildVersionString())
		os.Exit(1)
	}

	var dbWrapper t.Database

	dbPath, exists := os.LookupEnv("KP_DATABASE")
	if !exists {
		dbPath = *dbFile
	}

	// default to the flag argument
	keyPath := *keyFile

	if envKeyfile, found := os.LookupEnv("KP_KEYFILE"); found && keyPath == "" {
		keyPath = envKeyfile
	}

	for {
		// if the password is coming from an environment variable, we need to terminate
		// after the first attempt or it will fall into an infinite loop
		var err error
		password, passwordInEnv := os.LookupEnv("KP_PASSWORD")
		if !passwordInEnv {
			password, err = promptForDBPassword(shell)

			if err != nil {
				shell.Printf("could not retrieve password: %s", err)
				os.Exit(1)
			}
		}

		dbWrapper, err = newDB(dbPath, password, keyPath, *keepassVersion)
		if err != nil {
			// typically, these errors will be a bad password, so we want to keep prompting until the user gives up
			// if, however, the password is in an environment variable, we want to abort immediately so the program doesn't fall
			// in to an infinite loop
			shell.Printf("could not open database: %s\n", err)
			if passwordInEnv {
				os.Exit(1)
			}
			continue
		}
		break
	}

	if dbWrapper.Locked() {
		shell.Printf("Lockfile exists for DB at path '%s', another process is using this database!\n", dbWrapper.SavePath())
		shell.Printf("Open anyways? Data loss may occur. (will only proceed if 'yes' is entered)  ")
		line, err := shell.ReadLineErr()
		if err != nil {
			shell.Printf("could not read user input: %s\n", line)
			os.Exit(1)
		}

		if line != "yes" {
			shell.Println("aborting")
			os.Exit(1)
		}
	}

	if err := dbWrapper.Lock(); err != nil {
		shell.Printf("aborting, could not lock database: %s\n", err)
		os.Exit(1)
	}
	shell.Printf("opened database at %s\n", dbWrapper.SavePath())

	shell.Set("db", dbWrapper)
	shell.SetPrompt(fmt.Sprintf("/%s > ", dbWrapper.CurrentLocation().Name()))

	shell.AddCmd(&ishell.Cmd{
		Name:                "ls",
		Help:                "ls [path]",
		Func:                commands.Ls(shell),
		CompleterWithPrefix: fileCompleter(shell, true),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "new",
		Help:                "new <path>",
		LongHelp:            "creates a new entry at <path>",
		Func:                commands.NewEntry(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "mkdir",
		LongHelp:            "create a new group",
		Help:                "mkdir <group name>",
		Func:                commands.NewGroup(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:     "saveas",
		LongHelp: "save this db to a new file, existing credentials will be used, the new location will /not/ be used as the main save path",
		Help:     "saveas <file.kdb>",
		Func:     commands.SaveAs(shell),
	})

	if dbWrapper.Version() == t.V2 {
		shell.AddCmd(&ishell.Cmd{
			Name:                "select",
			Help:                "select [-f] <entry>",
			LongHelp:            "shows details on a given value in an entry, passwords will be redacted unless '-f' is specified",
			Func:                commands.Select(shell),
			CompleterWithPrefix: fileCompleter(shell, true),
		})
	}

	shell.AddCmd(&ishell.Cmd{
		Name:                "show",
		Help:                "show [-f] <entry>",
		LongHelp:            "shows details on a given entry, passwords will be redacted unless '-f' is specified",
		Func:                commands.Show(shell),
		CompleterWithPrefix: fileCompleter(shell, true),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "cd",
		Help:                "cd <path>",
		LongHelp:            "changes the current group to a different path",
		Func:                commands.Cd(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})

	attachCmd := &ishell.Cmd{
		Name:     "attach",
		LongHelp: "manages the attachment for a given entry",
		Help:     "attach <get|show|delete> <entry> <filesystem location>",
	}
	attachCmd.AddCmd(&ishell.Cmd{
		Name:                "create",
		Help:                "attach create <entry> <name> <filesystem location>",
		LongHelp:            "creates a new attachment based on a local file",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Attach(shell, "create"),
	})
	attachCmd.AddCmd(&ishell.Cmd{
		Name:                "get",
		Help:                "attach get <entry> <filesystem location>",
		LongHelp:            "retrieves an attachment and outputs it to a filesystem location",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Attach(shell, "get"),
	})
	attachCmd.AddCmd(&ishell.Cmd{
		Name:                "details",
		Help:                "attach details <entry>",
		LongHelp:            "shows the details of the attachment on an entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Attach(shell, "details"),
	})
	shell.AddCmd(attachCmd)

	shell.AddCmd(&ishell.Cmd{
		LongHelp:            "searches for any entries with the regular expression '<term>' in their titles or contents",
		Name:                "search",
		Help:                "search <term>",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Search(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "rm",
		Help:                "rm <entry>",
		LongHelp:            "removes an entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Rm(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xp",
		Help:                "xp <entry>",
		LongHelp:            "copies a password to the clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Xp(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "edit",
		Help:                "edit <entry>",
		LongHelp:            "edits an existing entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Edit(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "pwd",
		Help:     "pwd",
		LongHelp: "shows path of current group",
		Func:     commands.Pwd(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "save",
		Help:     "save",
		LongHelp: "saves the database to its most recently used path",
		Func:     commands.Save(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "xx",
		Help:     "xx",
		LongHelp: "clears the clipboard",
		Func:     commands.Xx(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xu",
		Help:                "xu",
		LongHelp:            "copies username to the clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Xu(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xw",
		Help:                "xw",
		LongHelp:            "copies url to clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Xw(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "mv",
		Help:                "mv <soruce> <destination>",
		LongHelp:            "moves entries between groups",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                commands.Mv(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "version",
		Help:     "version",
		LongHelp: "prints version",
		Func: func(c *ishell.Context) {
			shell.Printf("version: %s\n", buildVersionString())
		},
	})

	if *noninteractive != "" {
		bits := strings.Split(*noninteractive, " ")
		if err := shell.Process([]string{bits[0], strings.Join(bits[1:], " ")}...); err != nil {
			shell.Printf("error processing command: %s\n", err)
		}
	} else {
		shell.Run()
	}

	// This will run after the shell exits
	fmt.Println("exiting")

	if dbWrapper.Changed() {
		if err := commands.PromptAndSave(shell); err != nil {
			fmt.Printf("error attempting to save database: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("no changes detected since last save.")
	}

	if err := dbWrapper.Unlock(); err != nil {
		fmt.Printf("failed to unlock db: %s", err)
		os.Exit(1)
	}
}
