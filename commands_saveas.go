package main

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

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

// loadOrGenerateKey will return a file handle for the key, will prompt the user to generate a key if they so desire.
// if the key doesn't exist and the user declines to generate it, will return a nil reader and a nil error
// FIXME user should provide key, or utility should use actual key gen utility to do ti, this utility has no business generating one
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

		savePath := c.Args[0]
		if !confirmOverwrite(shell, savePath) {
			shell.Println("not overwriting existing file")
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

		db := shell.Get("db").(k.Database)
		shell.Printf("enter password: ")
		pw, err := shell.ReadPasswordErr()
		if err != nil {
			shell.Printf("could not read user input: %s\n", err)
			return
		}

		shell.Printf("enter password again: ")
		pwConfirm, err := shell.ReadPasswordErr()
		if err != nil {
			shell.Printf("could not read user input: %s\n", err)
			return
		}
		if pw != pwConfirm {
			shell.Println("password mismatch!")
			return
		}

		opts := k.Options{
			Password:  pw,
			KeyReader: file,
		}

		if err := db.SetOptions(opts); err != nil {
			shell.Printf("could not set db opts: %s", err)
			return
		}

		oldPath := db.SavePath()
		if err := removeLockfile(oldPath); err != nil {
			shell.Printf("could not remove lockfile from old path '%s'\n", oldPath)
		}
		db.SetSavePath(savePath)
		if err := db.Save(); err != nil {
			db.SetSavePath(oldPath)
			shell.Printf("could not save database: %s\n", err)
			// fall through to get the lock file back in place
		}

		if err := setLockfile(db.SavePath()); err != nil {
			shell.Printf("could not create lock file, data corruption may occur!: %s", err)
			return
		}
	}
}
