package keepassv1

import (
	"strings"
	"time"
	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// field name constants
const (
	fieldUn = "username"
	fieldPw = "password"
	fieldUrl = "url"
	fieldNotes = "notes"
	fieldTitle = "title"
	fieldAttachment = "attachment"
)

type Entry struct {
	entry *keepass.Entry
}

func NewEntry(entry *keepass.Entry) k.Entry {
	return &Entry{
		entry: entry,
	}
}

func (e *Entry) Title() string {
	return e.entry.Title
}

func (e *Entry) SetTitle(title string) {
	e.entry.Title = title
}

func (e *Entry) UUIDString() string {
	return e.entry.UUID.String()
}

func (e *Entry) Get(field string) (rv k.Value) {
	name := field
	var value interface{}
	switch strings.ToLower(field) {
	case fieldTitle:
		value = e.entry.Title
	case fieldUn:
		value = e.entry.Username
	case fieldPw:
		value = e.entry.Password
	case fieldUrl:
		value = e.entry.URL
	case fieldNotes:
		value = e.entry.Notes
	case fieldAttachment:
		if ! e.entry.HasAttachment() {
			return nil
		}
		value = e.entry.Attachment.Data
		name = e.entry.Attachment.Name
	default:
		return nil
	}
	return Value{
		name: name,
		value: value,
	}
}

func (e *Entry) Set(field string, value string) (updated bool) {
	updated = true
	switch strings.ToLower(field) {
	case fieldTitle:
		e.entry.Title = value
	case fieldUn:
		e.entry.Username = value
	case fieldPw:
		e.entry.Password = value
	case fieldUrl:
		e.entry.URL = value
	case fieldNotes:
		e.entry.Notes = value
	default:
		updated = false
	}
	return
}
func (e *Entry) Username() string {
	return e.entry.Username
}

func (e *Entry) Password() string {
	return e.entry.Password
}

func (e *Entry) URL() string {
	return e.entry.URL
}
func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.LastAccessTime = t
}

/**
// Copy returns a new copy of this wrapper, complete with a new keepass entry underneath it
// it also returns a boolean indicating whether the two entries differ
func Copy() (e Entry, changed bool) {
	changed = false
	if dest.Title != src.Title ||
		dest.Username != src.Username ||
		dest.Password != src.Password ||
		dest.Notes != src.Notes ||
		dest.URL != src.URL {
		changed = true
	}
	dest.Title = src.Title
	dest.Username = src.Username
	dest.Password = src.Password
	dest.Notes = src.Notes
	dest.URL = src.URL
	return changed
}
**/
