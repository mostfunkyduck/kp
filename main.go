package main

import (
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"log"
)

var (
	keyFile   = flag.String("key", "", "a key file to use to unlock the db")
	dbFile    = flag.String("db", "", "the db to open")
	debugMode = flag.Bool("debug", false, "verbose logging")
)

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
		Name: "ls",
		Help: "ls [path]",
		Func: Ls(shell),
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "show",
		Help: "show [-f] <entry>",
		Func: Show(shell),
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "cd <path>",
		Func: Cd(shell),
	})
	shell.Run()
}
