package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

var (
	keyFile   = flag.String("key", "", "a key file to use to unlock the db")
	dbFile    = flag.String("db", "", "the db to open")
	debugMode = flag.Bool("debug", false, "verbose logging")
	version   = flag.Bool("version", false, "print version and exit")
	DBChanged = false
)

func fileCompleter(shell *ishell.Shell, printEntries bool) func(string, []string) []string {
	return func(prefix string, args []string) (ret []string) {
		var rawPath string
		baseGroup := strings.Split(prefix, "/")
		baseGroup = baseGroup[0 : len(baseGroup)-1]
		rawPath = strings.Join(baseGroup, "/")

		location := shell.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(location, rawPath)
		if err != nil {
			return []string{}
		}

		if location != nil {
			if rawPath != "" {
				rawPath = rawPath + "/"
			}
			for _, g := range location.Groups() {
				ret = append(ret, rawPath+strings.ReplaceAll(g.Name, " ", "\\ ")+"/")
			}

			if printEntries {
				for _, e := range location.Entries() {
					ret = append(ret, rawPath+strings.ReplaceAll(e.Title, " ", "\\ "))
				}
			}
		}
		return ret
	}
}

func buildVersionString() string {
	return fmt.Sprintf("%s.%s-%s.%s (built on %s from %s)", VersionRelease, VersionBuildDate, VersionBuildTZ, VersionBranch, VersionHostname, VersionRevision)
}
func main() {
	flag.Parse()

	shell := ishell.New()
	if *version {
		shell.Printf("version: %s\n", buildVersionString())
		os.Exit(1)
	}

	shell.Set("filePath", *dbFile)
	var db *keepass.Database
	if *dbFile == "" {
		_db, err := keepass.New(&keepass.Options{})
		if err != nil {
			panic(err)
		}
		db = _db
	} else {
		_db, ok := openDB(shell)
		if !ok {
			shell.Println("could not open database")
			os.Exit(1)
		}
		db = _db
	}

	shell.Println("opened database")

	shell.Set("currentLocation", db.Root())
	shell.Set("db", db)
	shell.SetPrompt(fmt.Sprintf("%s > ", db.Root().Name))

	shell.AddCmd(&ishell.Cmd{
		Name:                "ls",
		Help:                "ls [path]",
		Func:                Ls(shell),
		CompleterWithPrefix: fileCompleter(shell, true),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "new",
		Help:                "new <path>",
		LongHelp:            "creates a new entry at <path>",
		Func:                NewEntry(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "mkdir",
		LongHelp:            "create a new group",
		Help:                "mkdir <group name>",
		Func:                NewGroup(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:     "saveas",
		LongHelp: "save this db to a new file with an optional key to be generated",
		Help:     "saveas <file.kdb> [file.key]",
		Func:     SaveAs(shell),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "show",
		Help:                "show [-f] <entry>",
		LongHelp:            "shows details on a given entry, passwords will be redacted unless '-f' is specified",
		Func:                Show(shell),
		CompleterWithPrefix: fileCompleter(shell, true),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "cd",
		Help:                "cd <path>",
		LongHelp:            "changes the current group to a different path",
		Func:                Cd(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})

	attachCmd := &ishell.Cmd{
		Name:     "attach",
		LongHelp: "manages the attachment for a given entry",
		Help:     "attach <get|show|delete> <entry> <filesystem location>",
	}
	attachCmd.AddCmd(&ishell.Cmd{
		Name:                "get",
		Help:                "attach get <entry> <filesystem location>",
		LongHelp:            "retrieves an attachment and outputs it to a filesystem location",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Attach(shell, "get"),
	})
	attachCmd.AddCmd(&ishell.Cmd{
		Name:                "details",
		Help:                "attach details <entry>",
		LongHelp:            "shows the details of the attachment on an entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Attach(shell, "details"),
	})
	shell.AddCmd(attachCmd)

	shell.AddCmd(&ishell.Cmd{
		LongHelp:            "searches for any entries with the regular expression '<term>' in their titles or contents",
		Name:                "search",
		Help:                "search <term>",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Search(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "rm",
		Help:                "rm <entry>",
		LongHelp:            "removes an entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Rm(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xp",
		Help:                "xp <entry>",
		LongHelp:            "copies a password to the clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Xp(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "edit",
		Help:                "edit <entry>",
		LongHelp:            "edits an existing entry",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Edit(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "pwd",
		Help:     "pwd",
		LongHelp: "shows path of current group",
		Func:     Pwd(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "save",
		Help:     "save",
		LongHelp: "saves the database to its most recently used path",
		Func:     Save(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "xx",
		Help:     "xx",
		LongHelp: "clears the clipboard",
		Func:     Xx(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xu",
		Help:                "xu",
		LongHelp:            "copies username to the clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Xu(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "xw",
		Help:                "xw",
		LongHelp:            "copies url to clipboard",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Xw(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:                "mv",
		Help:                "mv <soruce> <destination>",
		LongHelp:            "moves entries between groups",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Mv(shell),
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "version",
		Help:     "version",
		LongHelp: "prints version",
		Func: func(c *ishell.Context) {
			shell.Printf("version: %s\n", buildVersionString())
		},
	})

	shell.Run()

	// This will run after the shell exits
	if DBChanged {
		if err := promptAndSave(shell); err != nil {
			shell.Printf("error attempting to save database: %s\n", err)
		}
	}

	if err := removeLockfile(shell); err != nil {
		shell.Printf("could not remove lock file: %s\n", err)
	} else {
		shell.Println("no changes detected since last save.")
	}
}
