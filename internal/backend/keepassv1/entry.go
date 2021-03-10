package keepassv1

import (
	"fmt"
	"strings"
	"time"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// field name constants
const (
	fieldUn         = "username"
	fieldPw         = "password"
	fieldUrl        = "URL"
	fieldNotes      = "notes"
	fieldTitle      = "title"
	fieldAttachment = "attachment"
)

type Entry struct {
	c.Entry
	entry *keepass.Entry
}

func WrapEntry(entry *keepass.Entry, db t.Database) t.Entry {
	e := &Entry{
		entry: entry,
	}
	e.SetDB(db)
	e.SetDriver(e)
	return e
}

func (e *Entry) UUIDString() (string, error) {
	return e.entry.UUID.String(), nil
}

func (e *Entry) Get(field string) (rv t.Value) {
	var value []byte
	var name = field
	searchable := true
	protected := false
	valueType := t.STRING
	switch strings.ToLower(field) {
	case strings.ToLower(fieldTitle):
		value = []byte(e.entry.Title)
	case strings.ToLower(fieldUn):
		value = []byte(e.entry.Username)
	case strings.ToLower(fieldPw):
		searchable = false
		protected = true
		value = []byte(e.entry.Password)
	case strings.ToLower(fieldUrl):
		value = []byte(e.entry.URL)
	case strings.ToLower(fieldNotes):
		value = []byte(e.entry.Notes)
		valueType = t.LONGSTRING
	case strings.ToLower(fieldAttachment):
		if !e.entry.HasAttachment() {
			return nil
		}
		name = e.entry.Attachment.Name
		value = e.entry.Attachment.Data
		valueType = t.BINARY
	default:
		return nil
	}

	return c.NewValue(
		value,
		name,
		searchable,
		protected,
		false,
		valueType,
	)
}

func (e *Entry) Set(value t.Value) (updated bool) {
	updated = true
	field := value.Name()
	fieldValue := value.Value()
	switch strings.ToLower(field) {
	case strings.ToLower(fieldTitle):
		e.entry.Title = string(fieldValue)
	case strings.ToLower(fieldUn):
		e.entry.Username = string(fieldValue)
	case strings.ToLower(fieldPw):
		e.entry.Password = string(fieldValue)
	case strings.ToLower(fieldUrl):
		e.entry.URL = string(fieldValue)
	case strings.ToLower(fieldNotes):
		e.entry.Notes = string(fieldValue)
	case strings.ToLower(fieldAttachment):
		e.entry.Attachment.Name = field
		e.entry.Attachment.Data = fieldValue
	default:
		updated = false
	}
	return
}

func (e *Entry) LastAccessTime() time.Time {
	return e.entry.LastAccessTime
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.LastAccessTime = t
}

func (e *Entry) LastModificationTime() time.Time {
	return e.entry.LastModificationTime
}

func (e *Entry) SetLastModificationTime(t time.Time) {
	e.entry.LastModificationTime = t
}

func (e *Entry) CreationTime() time.Time {
	return e.entry.CreationTime
}

func (e *Entry) SetCreationTime(t time.Time) {
	e.entry.CreationTime = t
}

func (e *Entry) ExpiredTime() time.Time {
	return e.entry.ExpiryTime
}

func (e *Entry) SetExpiredTime(t time.Time) {
	e.entry.ExpiryTime = t
}
func (e *Entry) SetParent(g t.Group) error {
	if err := e.entry.SetParent(g.Raw().(*keepass.Group)); err != nil {
		return fmt.Errorf("could not set entry's group: %s", err)
	}
	return nil
}

func (e *Entry) Parent() t.Group {
	group := e.entry.Parent()
	if group == nil {
		return nil
	}
	return WrapGroup(group, e.DB())
}

func (e *Entry) Path() (string, error) {
	parent := e.Parent()
	if parent == nil {
		// orphaned entry
		return e.Title(), nil
	}
	groupPath, err := e.Parent().Path()
	if err != nil {
		return "", fmt.Errorf("could not find path to entry: %s", err)
	}
	return groupPath + e.Title(), nil
}

func (e *Entry) Raw() interface{} {
	return e.entry
}

func (e *Entry) Password() string {
	return string(e.Get("password").Value())
}

func (e *Entry) SetPassword(password string) {
	e.Set(c.NewValue(
		[]byte(password),
		"password",
		false,
		true,
		false,
		t.STRING,
	))
}

func (e *Entry) Title() string {
	return string(e.Get("title").Value())
}

func (e *Entry) SetTitle(title string) {
	e.Set(c.NewValue(
		[]byte(title),
		"title",
		true,
		false,
		false,
		t.STRING,
	))
}

func (e *Entry) Values() (vals []t.Value, err error) {
	path, _ := e.Path()
	vals = append(vals, c.NewValue([]byte(path), "location", false, false, true, t.STRING))
	vals = append(vals, e.Get(fieldTitle))
	vals = append(vals, e.Get(fieldUrl))
	vals = append(vals, e.Get(fieldUn))
	vals = append(vals, e.Get(fieldPw))
	vals = append(vals, e.Get(fieldNotes))
	if e.entry.HasAttachment() {
		vals = append(vals, e.Get(fieldAttachment))
	}
	return
}

func (e *Entry) Username() string {
	return e.Get(fieldUn).FormattedValue(true)
}

func (e *Entry) SetUsername(name string) {
	e.Set(c.NewValue(
		[]byte(name),
		fieldUn,
		true, false, false,
		t.STRING,
	))
}
