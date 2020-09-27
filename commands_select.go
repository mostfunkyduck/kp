package main

import (
	"strings"

	"github.com/abiosoft/ishell"
	// because ishell's checklist isn't rendering properly, at least on WSL
	"github.com/AlecAivazis/survey/v2"
)

func Select(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		// FIXME (medium priority) make this use a library for arg parsing so that we can have
		// it select fields inline
		fullMode := false
		path := c.Args[0]
		for _, arg := range c.Args {
			if strings.HasPrefix(arg, "-") {
				if arg == "-f" {
					fullMode = true
				}
				continue
			}
			path = arg
		}

		entry, ok := getEntryByPath(shell, path)
		if !ok {
			shell.Printf("could not retrieve entry at path '%s'\n", path)
			return
		}

		// now, prepare the checklist of fields to select

		// what the actual options are
		options := []string{}

		// what field names we want selected by default (case insensitive)
		defaultsRaw := []string{"password"}

		// what the actual defaults will be
		defaultSelections := []string{}
		for _, val := range entry.Values() {
			options = append(options, val.Name)
			for _, def := range defaultsRaw {
				if strings.EqualFold(def, val.Name) {
					defaultSelections = append(defaultSelections, val.Name)
				}
			}
		}
		selections := []string{}
		prompt := &survey.MultiSelect{
			VimMode: true, // duh
			Message: "Select fields to display",
			Options: options,
			Default: defaultSelections,
		}
		if err := survey.AskOne(prompt, &selections); err != nil {
			shell.Printf("could not select fields: %s", err)
			return
		}

		for _, val := range selections {
			fullValue := entry.Get(val)
			if fullValue.Name == "" {
				shell.Printf("error retrieving value for %s", fullValue.Value)
				return
			}

			value := string(fullValue.Value)
			if !fullMode && fullValue.Protected {
				value = "[protected]"
			}
			shell.Printf("%12s:\t%-12s\n", fullValue.Name, value)
		}
	}
}
