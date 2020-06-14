package main

import (
	"fmt"
	"strconv"
	"strings"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// manages the state of the shell
type ShellContext struct {
	CurrentLocation *keepass.Group
}

// flag contents for each command
type Flags map[string]string

// Commands execute actions, display usage strings, validate syntax.
// Output is sent to strings or errors for stdout/stderr
type Command interface {
	// Runs the command
	Execute(e *ShellContext) (string, error)

	// Display usage string
	Usage() string

	// Validates command arguments, if any exist
	Validate() error

	// Return the name of this command
	Name() string

	// Return the arguments, including flags for this command
	Arguments() []string

	// Return the flags for this command
	Flags() Flags
}

type command struct {
	// the actual command name
	name string

	// the arguments to the command
	arguments []string

	// the flags
	flags Flags
}

func (c command) Name() string {
	return c.name
}

func (c command) Arguments() []string {
	return c.arguments
}

func (c command) Flags() Flags {
	return c.flags
}

func GetCommand(cmd string, arguments []string) (Command, error) {
	flags := make(map[string]string)
	var actualArgs []string
	for _, arg := range arguments {
		if strings.HasPrefix(arg, "-") {
			flags[arg] = arg // may make this take actual argument at some point
		} else {
			actualArgs = append(actualArgs, arg)
		}
	}
	baseCommand := command{
		arguments: actualArgs,
		flags:     flags,
	}
	switch cmd {
	case "ls":
		return ls{baseCommand}, nil
	case "cd":
		return cd{baseCommand}, nil
	case "show":
		return show{baseCommand}, nil
	default:
		return nil, fmt.Errorf("command not found")
	}
}

type ls struct {
	command
}

func (l ls) Execute(s *ShellContext) (stdout string, stderr error) {
	// it appears that keepass uses null terminators at the end of its group and entries
	// lists, so i'm explicitly leaving those out of the output
	stdout = "groups:\n"
	for _, group := range s.CurrentLocation.Groups() {
		if group.Name != "" {
			stdout += fmt.Sprintf("%s/\n", group.Name)
		}
	}

	stdout += "entries:\n"
	for i, entry := range s.CurrentLocation.Entries() {
		if entry.Title != "" {
			stdout += fmt.Sprintf("%d: %s\n", i, entry.Title)
		}
	}
	return stdout, nil
}

func (l ls) Usage() string {
	return "ls [<group>]"
}

func (l ls) Validate() (err error) {
	if len(l.Arguments()) > 1 {
		err = fmt.Errorf("only one argument allowed")
	}
	return
}

type cd struct {
	command
}

func (c cd) Execute(s *ShellContext) (stdout string, stderr error) {
	args := c.Arguments()
	for _, group := range s.CurrentLocation.Groups() {
		if group.Name == args[0] {
			s.CurrentLocation = group
			return "", nil
		}
	}
	return "", fmt.Errorf("could not cd to %s", args[0])
}

func (c cd) Usage() string {
	return "cd <group>"
}

func (c cd) Validate() (err error) {
	if len(c.Arguments()) != 1 {
		err = fmt.Errorf("one and only one argument allowed")
	}
	return
}

type show struct {
	command
}

func (s show) Execute(sh *ShellContext) (stdout string, stderr error) {
	args := s.Arguments()
	flags := s.Flags()
	entryIndex, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("%s is not a valid index", args[0])
	}

	entries := sh.CurrentLocation.Entries()
	if len(entries)-1 < entryIndex {
		return "", fmt.Errorf("invalid entry index %d", entryIndex)
	}

	for i, entry := range entries {
		if i == entryIndex {
			stdout += fmt.Sprintf("Title: %s\n", entry.Title)
			stdout += fmt.Sprintf("URL: %s\n", entry.URL)
			stdout += fmt.Sprintf("Username: %s\n", entry.Username)
			stdout += "Password: "
			if _, ok := flags["-f"]; ok {
				stdout += entry.Password
			}
			stdout += "\n"
			stdout += fmt.Sprintf("Notes: %s\n", entry.Notes)
			stdout += "Attachment: "
			if entry.HasAttachment() {
				stdout += entry.Attachment.Name
			}
			stdout += "\n"
		}
	}
	return stdout, nil
}

func (s show) Usage() string {
	return "show [-f] <entry>"
}

func (s show) Validate() (err error) {
	if len(s.Arguments()) < 1 {
		err = fmt.Errorf("at least one argument required")
	}
	return
}
