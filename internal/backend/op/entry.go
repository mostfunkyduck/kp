package op
import (
	"fmt"
	"strings"
	"time"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

type Entry struct {
	c.Entry
	UUID			string
	db				t.Database
	// contents of the entry as stored in the 'list item' call - mostly metadata
	listItem	opListItem
	// contents of the entry as stored in the 'get item' call - mostly details
	getItem		opGetItem
	parent		t.Group
	opDriver	*Driver
}

// Init does the internal tricks to get entries to work properly within the system
func (e *Entry) Init() error {
	e.SetDriver(e)
	return nil
}

func (e Entry) UUIDString() (string, error) {
	return e.UUID, nil
}

func (e Entry) Raw() interface{} {
	return e
}

// NOTE: this will pull the live data as that's the only place you get passwords
func (e Entry) Password() string {
	e.pullLive()
	return string(e.Get("password").Value())
}

func (e Entry) SetPassword(p string) {
	// only read only supported rn
	return
}

// getField is a helper function for retrieving fields from a getItem's details.fields array
func (e Entry) getItemField(fieldName string) string {
	for _, f := range e.getItem.Details.Fields {
		if strings.ToLower(f.Name) == strings.ToLower(fieldName) {
			return f.Value
		}
	}
	return "undefined"
}

// pullLive uses the 'get item' command to get the full details of the entry
func (e *Entry) pullLive() error {
	liveEntry, err := e.opDriver.GetItem(e.UUID)
	if err != nil {
		return fmt.Errorf("could not get live copy of entry: %s", err)
	}
	e.getItem = liveEntry.getItem
	return nil
}

// TODO when creation is implemented, will need to handle local vs. remote
func (e Entry) Get(field string) t.Value {
	protected := false
	val := "undefined"
	switch field {
	case strings.ToLower("title"):
		val = "undefined"
		if e.getItem.Overview.Title != "" {
			val = e.getItem.Overview.Title
		} else if e.listItem.Overview.Title != "" {
			val = e.listItem.Overview.Title
		}
	case strings.ToLower("username"):
		val = e.getItemField(field)
	case strings.ToLower("password"):
		val = e.getItemField(field)
		protected = true
	case strings.ToLower("notes"):
		val = e.getItem.Details.Notes
	case strings.ToLower("ainfo"):
		val = e.getItem.Overview.Ainfo
	}
	return c.NewValue (
		[]byte(val),
		field,
		true,
		protected,
		false,
		t.STRING,
	)
}


func (e *Entry) Set(value t.Value) bool {
	// not implemented
	return true
}

func (e *Entry) SetTitle(title string) {
	// only read only supported
}

func (e Entry) Title() string {
	return string(e.Get("title").Value())
}

func (e Entry) LastAccessTime() time.Time {
	return time.Time{}
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	// 1pass doesn't track this, no-op
}

func (e Entry) LastModificationTime() time.Time {
	return e.listItem.UpdatedAt
}

func (e *Entry) SetLastModificationTime(t time.Time) {
	// not implemented yet
}

func (e Entry) CreationTime() time.Time {
	return e.listItem.CreatedAt
}

func (e *Entry) SetCreationTime(t time.Time) {
	// not implemented yet
}

// I don't see 1pass returning this
func (e Entry) ExpiredTime() time.Time {
	return time.Time{}
}

// same, 1pass isn't doing it, so we don't
func (e Entry) SetExpiredTime(t time.Time) {
	return
}

func (e Entry) Parent() t.Group {
	return e.parent
}

func (e Entry) SetParent(g t.Group) error {
	e.parent = g
	return nil
}

// Values in this backed will pull live data to ensure accuracy and completeness
// It CANNOT be called simultaneously on large lists of entries (i.e "ls") without bombarding the 1pass server
func (e Entry) Values() (values []t.Value, err error) {
	if err := e.pullLive(); err != nil {
		return []t.Value{}, err
	}
	// Get has to be called directly because the helpers (e.g Title), don't return Values
	values = append(values, e.Get("title"))
	values = append(values, e.Get("password"))
	values = append(values, e.Get("ainfo"))
	return
}

// FIXME this should be in the abstract class....
func (e Entry) DB() t.Database {
	return e.db
}

func (e *Entry) SetDB(db t.Database) {
	e.db = db
}

func (e *Entry) SetUsername(un string) {
	// not implemented yet
}

func (e *Entry) Username() string {
	return string(e.Get("username").Value())
}


