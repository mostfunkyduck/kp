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

func (e *Entry) Get(field string) (rv t.Value, present bool) {
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
			return nil, false
		}
		return c.Attachment{
			EntryValue: c.NewValue(
				e.entry.Attachment.Data,
				e.entry.Attachment.Name,
				searchable,
				protected,
				false,
				t.BINARY,
			),
		}, true
	default:
		return nil, false
	}

	return c.NewValue(
		value,
		name,
		searchable,
		protected,
		false,
		valueType,
	), true
}

func (e *Entry) Set(value t.Value) (updated bool) {
	updated = true
	field := value.Name()
	fieldValue := value.Value()

	if value.Type() == t.BINARY {
		e.entry.Attachment.Name = field
		e.entry.Attachment.Data = fieldValue
		return true
	}

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
	v, _ := e.Get("password")
	return string(v.Value())
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
	v, _ := e.Get("title")
	return string(v.Value())
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
	for _, field := range []string{fieldTitle, fieldUrl, fieldUn, fieldPw, fieldNotes, fieldAttachment} {
		if v, present := e.Get(field); present {
			vals = append(vals, v)
		}
	}
	return
}

func (e *Entry) Username() string {
	v, _ := e.Get(fieldUn)
	return v.FormattedValue(true)
}

func (e *Entry) SetUsername(name string) {
	e.Set(c.NewValue(
		[]byte(name),
		fieldUn,
		true, false, false,
		t.STRING,
	))
}
