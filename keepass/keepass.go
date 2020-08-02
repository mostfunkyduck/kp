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
	// Raw returns the underlying object that the wrapper wraps aroud
	Raw() interface{}

	// Path returns the path to the object's location
	Path() (string, error)

	// Search searches this object and all nested objects for a given regular expression
	Search(*regexp.Regexp) []string
}

type Database interface {
	KeepassWrapper
	// CurrentLocation returns the current location for the shell
	CurrentLocation() Group
	SetCurrentLocation(Group)
	Root() Group
	Save() error

	// SavePath returns the path to which the database will be saved
	SavePath() string
	SetSavePath(newPath string)

	// SetOptions sets options for interacting with the database file
	SetOptions(Options) error
}

// Options for SetOptions in the database interface
type Options struct {
	KeyReader io.Reader
	Password  string
}

type UUIDer interface {
	// UUIDString returns the string form of this object's UUID
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
}

type Entry interface {
	UUIDer
	KeepassWrapper
	// Returns the value for a given field, or nil if the field doesn't exist
	Get(string) Value

	// Title and Password are needed to ensure that v1 and v2 both render
	// their specific representations of that data (they access it in different ways, fun times)
	Title() string
	SetTitle(string)
	Password() string
	SetPassword(string)

	// Sets a given field to a given value, returns bool indicating whether or not the field was updated
	Set(value Value) bool

	LastAccessTime() time.Time
	SetLastAccessTime(time.Time)

	LastModificationTime() time.Time
	SetLastModificationTime(time.Time)

	CreationTime() time.Time
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

	// DB returns the Database this entry is associated with
	DB() Database
	SetDB(Database)
}

type Value struct {
	Value      interface{} // can be either binary or string data
	Name       string      // v1 compatibility - attachments have their own name within entries
	Protected  bool        // only useable in v2, whether the value should be encrypted
	Searchable bool        // indicates whether this value should be included in searches
}
