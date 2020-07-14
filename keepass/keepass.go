package keepass

import (
	"io"
	"time"
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

type Group interface {
	// Returns all entries in this group
	Entries() []Entry
	// Returns all groups nested in this group
	Groups() []Group
	Parent() Group
	Name() string
	IsRoot() bool
}

type Entry interface {
	Title() string
	// We only need the string version of the UUID for this application
	UUIDString() string
	// Returns the value for a given field, or "" if the field doesn't exist
	Get(string) Value

	// Sets a given field to a given value, returns bool indicating whether or not the field was updated
	Set(field string, value string) bool

	// Sets the last accessed time on the entry
	SetLastAccessTime(time.Time)
}

type Value interface {
	Name() string
	Value() interface{}
}
