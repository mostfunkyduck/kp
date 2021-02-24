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
	Search(*regexp.Regexp) ([]string, error)
}

type Database interface {
	KeepassWrapper
	// Binary returns a binary with a given ID, naming it with a given name
	// the OptionalWrapper is used because v2 is the only version that implements this
	Binary(id int, name string) (OptionalWrapper, error)
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
	// Returns the value for a given field, or an empty struct if the field doesn't exist
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

	ExpiredTime() time.Time
	SetExpiredTime(time.Time)

	Parent() Group
	SetParent(Group) error

	// Formats an entry for printing
	Output(full bool) string

	// Values returns all referencable value fields from the database
	//
	// NOTE: values are not references, updating them must be done through the Set* functions
	Values() (values []Value, err error)

	// DB returns the Database this entry is associated with
	DB() Database
	SetDB(Database)
}

type ValueType int

const (
	STRING ValueType = iota
	LONGSTRING
	BINARY
)

// OptionalWrapper wraps Values with functions that force the caller of a function to detect whether the value being
// returned is implemented by the function, this is to help bridge the gap between v2 and v1
// Proper usage:
// if wrapper.Present {
//   <use value>
// } else {
// 	 <adapt>
// }
type OptionalWrapper struct {
	Present bool
	Value   Value
}

type Value interface {
	FormattedValue(full bool) string
	Value() []byte
	Name() string
	Searchable() bool
	Protected() bool
	ReadOnly() bool
	Type() ValueType
}
