package main

import (
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

var (
	keyFile = flag.String("key", "", "a key file to use to unlock the db")
	dbFile  = flag.String("db", "", "the db to open")
)

func main() {
	flag.Parse()

	if *dbFile == "" {
		log.Fatalf("no db file provided!")
	}

	dbReader, err := os.Open(*dbFile)
	if err != nil {
		log.Fatalf("could not open db file [%s]: %s", *dbFile, err)
	}

	var keyReader io.Reader
	if *keyFile != "" {
		keyReader, err = os.Open(*keyFile)
		if err != nil {
			log.Fatalf("could not open key file %s", *keyFile)
		}
	}
	shell := ishell.New()

	shell.Println("enter database password")
	password, err := shell.ReadPasswordErr()
	if err != nil {
		log.Fatalf("could not obtain password: %s", password)
	}

	opts := &keepass.Options{
		Password: password,
		KeyFile:  keyReader,
	}

	db, err := keepass.Open(dbReader, opts)
	if err != nil {
		log.Fatalf("could not open database [%s]: %s", *dbFile, err)
	}

	shell.Println("opened database")
	shell.Set("currentLocation", db.Root())
	shell.SetPrompt(fmt.Sprintf("%s > ", db.Root().Name))
	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "show entries in group",
		Func: func(c *ishell.Context) {
			currentLocation := c.Get("currentLocation").(*keepass.Group)
			lines := []string{}
			for _, group := range currentLocation.Groups() {
				lines = append(lines, fmt.Sprintf("%s/", group.Name))
			}
			for i, entry := range currentLocation.Entries() {
				lines = append(lines, fmt.Sprintf("%d: %s", i, entry.Title))
			}
			c.Println(strings.Join(lines, "\n"))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "show",
		Help: "show [-f] <entry>",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("incorrect number of arguments to show"))
				return
			}

			fullMode := false
			entryName := c.Args[0]
			if c.Args[0] == "-f" {
				fullMode = true
				if len(c.Args) != 2 {
					c.Err(fmt.Errorf("no second argument to show"))
				}
				entryName = c.Args[1]
			}
			currentLocation := c.Get("currentLocation").(*keepass.Group)
			for i, entry := range currentLocation.Entries() {
				if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i {
					outputEntry(*entry, c, fullMode)
					break
				}

				if entryName == entry.Title {
					outputEntry(*entry, c, fullMode)
					break
				}
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "change current group",
		Func: func(c *ishell.Context) {
			args := c.Args
			currentLocation := c.Get("currentLocation").(*keepass.Group)
			update := false
			defer func() {
				if update {
					shell.Set("currentLocation", currentLocation)
					c.SetPrompt(fmt.Sprintf("%s > ", currentLocation.Name))
				} else {
					c.Err(fmt.Errorf("invalid group"))
				}
			}()
			if len(args) == 0 {
				currentLocation = db.Root()
				update = true
				return
			}
			path := strings.Split(args[0], "/")

			for _, part := range path {
				if part == "." {
					continue
				}

				if part == ".." {
					if currentLocation.Parent() != nil {
						currentLocation = currentLocation.Parent()
						update = true
						continue
					}
				} else {
					for _, group := range currentLocation.Groups() {
						if group.Name == part {
							currentLocation = group
							update = true
							break
						}
					}
				}
			}
		}})
	shell.Run()
}

func outputEntry(e keepass.Entry, c *ishell.Context, full bool) {
	c.Println(fmt.Sprintf("Title: %s", e.Title))
	c.Println(fmt.Sprintf("URL: %s", e.URL))
	c.Println(fmt.Sprintf("Username: %s", e.URL))
	password := "[redacted]"
	if full {
		password = e.Password
	}
	c.Println(fmt.Sprintf("Password: %s", password))
	c.Println(fmt.Sprintf("Notes : %s", e.Notes))
	if e.HasAttachment() {
		c.Println(fmt.Sprintf("Attachment: %s", e.Attachment.Name))
	}

}
