package main

import (
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"log"
	"strings"
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

	db, ok := openDB(shell)
	if !ok {
		log.Fatalf("could not open database")
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
	shell.AddCmd(&ishell.Cmd{
		Name: "attach",
		Help: "attach <get|show|delete> <entry> <filesystem location>",
		Func: Attach(shell),
	})
	shell.Run()
}
