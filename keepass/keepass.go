package keepass

import (
	"io"
	"regexp"
	"time"
)

type version int

const (
	V1 version = iota
	V2
)

// abstracts a wrapper for v1 or v2 implementations to use to describe the database and to implement shell commands
type KeepassWrapper interface {
	// Returns the underlying object that the wrapper wraps aroud
	Raw() interface{}

	// returns the path to the object's location
	Path() (string, error)
}

type Database interface {
	KeepassWrapper
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
}

// Options for SetOptions in the database interface
type Options struct {
	KeyReader io.Reader
	Password  string
}

type UUIDer interface {
	// We only need the string version of the UUID for this application
	UUIDString() (string, error)
}
type Group interface {
	KeepassWrapper
	UUIDer
	// Returns all entries in this group
	Entries() []Entry

	// Returns all groups nested in this group
	Groups() []Group

	// Returns this group's parent, if it has one
	Parent() Group
	SetParent(Group) error
	// inverse of 'SetParent', needed mainly for the internals of keepassv2
	AddEntry(Entry) error

	Name() string
	SetName(string)

	IsRoot() bool

	// Creates a new subgroup with a given name under this group
	NewSubgroup(name string) (Group, error)
	RemoveSubgroup(Group) error
	AddSubgroup(Group) error

	NewEntry(name string) (Entry, error)
	RemoveEntry(Entry) error

	Search(*regexp.Regexp) []string
}

type Entry interface {
	UUIDer
	KeepassWrapper
	// Returns the value for a given field, or nil if the field doesn't exist
	Get(string) Value

	// Title and Password are needed to ensure that v1 and v2 both render
	// their specific representations of that data (they access it in different ways, fun times)
	Title() string
	Password() string

	// Sets a given field to a given value, returns bool indicating whether or not the field was updated
	Set(value Value) bool

	// Sets the last accessed time on the entry
	SetLastAccessTime(time.Time)
	SetLastModificationTime(time.Time)
	SetCreationTime(time.Time)

	Parent() Group
	SetParent(Group) error

	// Formats an entry for printing
	Output(full bool) string

	// Values returns all referencable value fields from the database
	//
	// NOTE: in keepass 1, this means that the hardcoded fields
	// will be returned in the Value wrapper.
	//
	// NOTE: values are read only
	Values() (values []Value)
}

type Value struct {
	Value     interface{} // can be either binary or string data
	Name      string      // v1 compatibility - attachments have their own name within entries
	Protected bool        // only useable in v2, whether the value should be encrypted
}
