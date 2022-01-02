package types

import (
	"regexp"
	"time"
)

type Version int

const (
	V1 Version = iota
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

type Backend interface {
	// Filename returns the file name for the backend storage
	Filename() string

	// Hash returns the cached hash representing a unique state of the backend storage
	Hash() string

	// IsModified returns whether or not the backend has been modified since it was last hashed
	IsModified() (bool, error)
}

type Database interface {
	KeepassWrapper
	// Backend returns the functions backend struct
	Backend() Backend

	// Binary returns a binary with a given ID, naming it with a given name
	// the OptionalWrapper is used because v2 is the only version that implements this
	Binary(id int, name string) (OptionalWrapper, error)

	// Changed indicates whether the DB has been changed during the user's session
	Changed() bool
	SetChanged(bool)

	// CurrentLocation returns the current location for the shell
	CurrentLocation() Group
	SetCurrentLocation(Group)
	Root() Group
	Save() error

	// Init initializes a database wrapper, using the given parameters.  Existing DB will be opened, otherwise the wrapper will be configured to save to that location
	Init(Options) error

	// Lock will lock the database by dropping a lockfile
	Lock() error

	// Unlock will remove the lockfile created by Lock()
	Unlock() error

	// Locked will determine if the lockfile is in place
	Locked() bool

	// SavePath and SetSavePath are shortcuts for managing the backend filename
	SavePath() string
	SetSavePath(string)

	// Version will return the Version enum for this database
	Version() Version
}

// Options are parameters to use for calls to the database interface's Init function
type Options struct {
	// the path to the database
	DBPath string

	// the path to the key
	KeyPath string

	// the password for the database
	Password string

	// How many rounds of encryption to use for the new key (currently only supported by keepassv1)
	KeyRounds int
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

	Username() string
	SetUsername(name string)

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
