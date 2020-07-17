package keepassv1

import (
	"fmt"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	k "github.com/mostfunkyduck/kp/keepass"
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
	entry *keepass.Entry
}

func NewEntry(entry *keepass.Entry) k.Entry {
	return &Entry{
		entry: entry,
	}
}

func (e *Entry) UUIDString() string {
	return e.entry.UUID.String()
}

func (e *Entry) Get(field string) (rv k.Value) {
	name := field
	var value string
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
		if !e.entry.HasAttachment() {
			return k.Value{}
		}
		return k.Value{
			Name:  e.entry.Attachment.Name,
			Value: e.entry.Attachment.Data,
		}
	default:
		return k.Value{}
	}
	return k.Value{
		Name:  name,
		Value: value,
	}
}

func (e *Entry) Set(field string, value k.Value) (updated bool) {
	updated = true
	switch strings.ToLower(field) {
	case fieldTitle:
		e.entry.Title = value.Value.(string)
	case fieldUn:
		e.entry.Username = value.Value.(string)
	case fieldPw:
		e.entry.Password = value.Value.(string)
	case fieldUrl:
		e.entry.URL = value.Value.(string)
	case fieldNotes:
		e.entry.Notes = value.Value.(string)
	case fieldAttachment:
		e.entry.Attachment.Name = value.Name
		e.entry.Attachment.Data = value.Value.([]byte)
	default:
		updated = false
	}
	return
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.LastAccessTime = t
}

func (e *Entry) SetLastModificationTime(t time.Time) {
	e.entry.LastModificationTime = t
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
	return &Group{
		group: e.entry.Parent(),
	}
}

func (e *Entry) Path() string {
	groupPath := e.Parent().Path()
	return groupPath + e.Get("title").Value.(string)
}

func (e *Entry) Raw() interface{} {
	return e.entry
}

func formatTime(t time.Time) (formatted string) {
	timeFormat := "Mon Jan 2 15:04:05 MST 2006"
	if (t == time.Time{}) {
		formatted = "unknown"
	} else {
		since := time.Since(t).Round(time.Duration(1) * time.Second)
		sinceString := since.String()

		// greater than or equal to 1 day
		if since.Hours() >= 24 {
			sinceString = fmt.Sprintf("%d days ago", int(since.Hours()/24))
		}

		// greater than or equal to ~1 month
		if since.Hours() >= 720 {
			// rough estimate, not accounting for non-30-day months
			months := int(since.Hours() / 720)
			sinceString = fmt.Sprintf("about %d months ago", months)
		}

		// greater or equal to 1 year
		if since.Hours() >= 8760 {
			// yes yes yes, leap years aren't 365 days long
			years := int(since.Hours() / 8760)
			sinceString = fmt.Sprintf("about %d years ago", years)
		}

		// less than a second
		if since.Seconds() < 1.0 {
			sinceString = "less than a second ago"
		}

		formatted = fmt.Sprintf("%s (%s)", t.Local().Format(timeFormat), sinceString)
	}
	return
}

func (e *Entry) Output(s *ishell.Shell, full bool) {
	s.Printf("\n")
	s.Printf("UUID:\t%s\n", e.entry.UUID)

	s.Printf("Creation Time:\t%s\n", formatTime(e.entry.CreationTime))
	s.Printf("Last Modified:\t%s\n", formatTime(e.entry.LastModificationTime))
	s.Printf("Last Accessed:\t%s\n", formatTime(e.entry.LastAccessTime))
	s.Printf("Location:\t%s\n", e.Path())
	s.Printf("Title:\t%s\n", e.Get("title").Value.(string))
	s.Printf("URL:\t%s\n", e.Get("url").Value.(string))
	s.Printf("Username:\t%s\n", e.Get("username").Value.(string))
	password := "[redacted]"
	if full {
		password = e.Get("password").Value.(string)
	}
	s.Printf("Password:\t%s\n", password)
	s.Printf("Notes:\n%s\n", e.Get("notes").Value.(string))
	if e.entry.HasAttachment() {
		s.Printf("Attachment:\t%s\n", e.Get("attachment").Name)
	}
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
