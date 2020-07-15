package main

import (
	"regexp"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
)

// FIXME the keepass library has a bug where you can't get the parent
// unless the entry is a pointer to the one in the db (it's comparing pointer values)
// this can/should/will be fixed in my fork
func searchEntries(g k.Group, term *regexp.Regexp) (titles []string) {
	for _, e := range g.Entries() {
		if term.FindString(e.Get("title").Value.(string)) != "" ||
			term.FindString(e.Get("notes").Value.(string)) != "" ||
			term.FindString(e.Get("attachment").Name) != "" ||
			term.FindString(e.Get("username").Value.(string)) != "" {
			titles = append(titles, e.Get("title").Value.(string))
		}
	}
	return titles
}

// searchGroup recursively searches groups and entries in a given group matching a given term
func searchGroup(g k.Group, term *regexp.Regexp, path string) (paths []string) {
	// the initial group will send in "", meaning it should be skipped in the path output
	if path != "" {
		path = path + "/" + g.Name()
		if term.FindString(g.Name()) != "" {
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
		paths = append(paths, searchGroup(g, term, path)...)
	}
	return paths
}

// This implements the equivalent of kpcli's "find" command, just with a name
// that won't be confused for the shell command of the same name
func Search(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		currentLocation := shell.Get("db").(k.Database).Root()
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
		for _, result := range searchGroup(currentLocation, term, "") {
			// the tab makes it a little more readable
			shell.Printf("\t%s\n", result)
		}
	}
}
