package keepassv1

import (
	"fmt"
	"strings"
	"time"

	k "github.com/mostfunkyduck/kp/keepass"
	c "github.com/mostfunkyduck/kp/keepass/common"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// field name constants
const (
	fieldUn         = "username"
	fieldPw         = "password"
	fieldUrl        = "url"
	fieldNotes      = "notes"
	fieldTitle      = "title"
	fieldAttachment = "attachment"
)

type Entry struct {
	c.Entry
	entry *keepass.Entry
}

func WrapEntry(entry *keepass.Entry, db k.Database) k.Entry {
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

func (e *Entry) Get(field string) (rv k.Value) {
	switch strings.ToLower(field) {
	case fieldTitle:
		rv.Value = []byte(e.entry.Title)
	case fieldUn:
		rv.Value = []byte(e.entry.Username)
	case fieldPw:
		rv.Value = []byte(e.entry.Password)
	case fieldUrl:
		rv.Value = []byte(e.entry.URL)
	case fieldNotes:
		rv.Value = []byte(e.entry.Notes)
	case fieldAttachment:
		if !e.entry.HasAttachment() {
			return k.Value{}
		}
		return k.Value{
			Name:  e.entry.Attachment.Name,
			Value: e.entry.Attachment.Data,
		}
	}
	if string(rv.Value) != "" {
		rv.Name = field
	}

	return
}

func (e *Entry) Set(value k.Value) (updated bool) {
	updated = true
	field := value.Name
	switch strings.ToLower(field) {
	case fieldTitle:
		e.entry.Title = string(value.Value)
	case fieldUn:
		e.entry.Username = string(value.Value)
	case fieldPw:
		e.entry.Password = string(value.Value)
	case fieldUrl:
		e.entry.URL = string(value.Value)
	case fieldNotes:
		e.entry.Notes = string(value.Value)
	case fieldAttachment:
		e.entry.Attachment.Name = value.Name
		e.entry.Attachment.Data = value.Value
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

func (e *Entry) SetParent(g k.Group) error {
	if err := e.entry.SetParent(g.Raw().(*keepass.Group)); err != nil {
		return fmt.Errorf("could not set entry's group: %s", err)
	}
	return nil
}

func (e *Entry) Parent() k.Group {
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
	return string(e.Get("password").Value)
}

func (e *Entry) SetPassword(password string) {
	e.Set(k.Value{Name: "password", Value: []byte(password)})
}

func (e *Entry) Title() string {
	return string(e.Get("title").Value)
}

func (e *Entry) SetTitle(title string) {
	e.Set(k.Value{Name: "title", Value: []byte(title)})
}

func (e *Entry) Values() (vals []k.Value, err error) {
	path, _ := e.Path()
	vals = append(vals, k.Value{Name: "location", Value: []byte(path), ReadOnly: true, Searchable: false})
	vals = append(vals, k.Value{Name: fieldTitle, Value: []byte(e.Title()), Searchable: true})
	vals = append(vals, k.Value{Name: "URL", Value: []byte(e.Get(fieldUrl).Value), Searchable: true})
	vals = append(vals, k.Value{Name: fieldUn, Value: []byte(e.Get(fieldUn).Value), Searchable: true})
	vals = append(vals, k.Value{Name: fieldPw, Value: []byte(e.Password()), Searchable: true, Protected: true})
	vals = append(vals, k.Value{Name: fieldNotes, Value: []byte(e.Get(fieldNotes).Value), Searchable: true, Type: k.LONGSTRING})
	if e.entry.HasAttachment() {
		vals = append(vals, k.Value{Name: fieldAttachment, Value: e.Get(fieldAttachment).Value, Searchable: true, Type: k.BINARY})
	}
	return
}
