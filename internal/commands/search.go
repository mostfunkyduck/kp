package commands

import (
	"regexp"

	"github.com/mostfunkyduck/ishell"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

// This implements the equivalent of kpcli's "find" command, just with a name
// that won't be confused for the shell command of the same name
func Search(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := shell.Get("db").(t.Database).Root()
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
		searchResults, err := currentLocation.Search(term)
		if err != nil {
			shell.Println("error during search: " + err.Error())
			return
		}
		for _, result := range searchResults {
			// the tab makes it a little more readable
			shell.Printf("\t%s\n", result)
		}
	}
}
