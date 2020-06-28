package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

var (
	keyFile   = flag.String("key", "", "a key file to use to unlock the db")
	dbFile    = flag.String("db", "", "the db to open")
	debugMode = flag.Bool("debug", false, "verbose logging")
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
				ret = append(ret, rawPath+g.Name+"/")
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

func main() {
	flag.Parse()

	shell := ishell.New()
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
			log.Fatalf("could not open database")
		}
		db = _db
	}

	shell.Println("opened database")

	shell.Set("currentLocation", db.Root())
	shell.Set("db", db)
	shell.Set("filePath", *dbFile)
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
		Func:                Show(shell),
		CompleterWithPrefix: fileCompleter(shell, true),
	})
	shell.AddCmd(&ishell.Cmd{
		Name:                "cd",
		Help:                "cd <path>",
		Func:                Cd(shell),
		CompleterWithPrefix: fileCompleter(shell, false),
	})

	attachCmd := &ishell.Cmd{
		Name: "attach",
		Help: "attach <get|show|delete> <entry> <filesystem location>",
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
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Attach(shell, "details"),
	})
	shell.AddCmd(attachCmd)

	shell.AddCmd(&ishell.Cmd{
		Name:                "search",
		Help:                "search <term>",
		CompleterWithPrefix: fileCompleter(shell, true),
		Func:                Search(shell),
	})

	shell.Run()
}
