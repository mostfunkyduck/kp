package keepass

import (
	"io"
)

type version int
const (
	V1 version = iota
	V2
)

// abstracts a wrapper for v1 or v2 implementations to use to describe the database and to implement shell commands
type Database interface {
	// Returns the current location for the shell
	CurrentLocation() Group
	// Returns the root of the database
	Root() Group
	Save() error
	// Returns the path of the DB on the filesystem
	SavePath() string
	SetCurrentLocation(Group)
	SetSavePath(newPath string)
	// Sets options for interacting with the database file
	SetOptions(Options) error
	// Navigates a path starting from the Group provided
	TraversePath(Group, string) (Group, error)
}

// Options for SetOptions in the database interface
type Options struct {
	KeyReader	io.Reader
	Password	string
}

type Commands interface {
	Cd() error
	Save() error
}

type Group interface {
	// Returns all entries in this group
	Entries() []Entry
	// Returns all groups nested in this group
	Groups() []Group
	Parent() Group
	Name() string
}

type Entry interface {
	Title() string
}

