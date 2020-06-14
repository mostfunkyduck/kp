package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"bufio"
	"log"
	"os"
	"strings"
	"syscall"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

var (
	keyFile = flag.String("key", "", "a key file to use to unlock the db")
	dbFile	= flag.String("db", "", "the db to open")
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

	password, err := getPassword()
	if err != nil {
		log.Fatalf("could not obtain password: %s", password)
	}

	opts := &keepass.Options{
		Password: password,
		KeyFile: keyReader,
	}

	db, err := keepass.Open(dbReader, opts)
	if err != nil {
		log.Fatalf("could not open database [%s]: %s", *dbFile, err)
	}
	fmt.Println("opened database!")
	shell(db)
}

func shell(db *keepass.Database) {
	stdinReader := bufio.NewReader(os.Stdin)
	ctx := &ShellContext{
		CurrentLocation: db.Root(),
	}
	for {
		fmt.Printf("%s > ", ctx.CurrentLocation.Name)
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Printf("could not read input: %s\n", err)
			os.Exit(1)
		}

		input = strings.Trim(input, "\n")
		inputComponents := strings.Split(input, " ")
		var args []string
		if len(inputComponents) > 1 {
			args = inputComponents[1:]
		}
		command, err := GetCommand(inputComponents[0], args)
		if err != nil {
			fmt.Printf("%s: %s\n", input, err)
			continue
		}

		if validationError := command.Validate(); validationError != nil {
			fmt.Printf("%s: %s\n", command.Name(), validationError)
			continue
		}

		output, err := command.Execute(ctx)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}
		fmt.Println(output)
	}
}
func getPassword() (password string, err error) {
	fmt.Printf("Enter password:\n")
	//https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	return string(passwordBytes), err
}
