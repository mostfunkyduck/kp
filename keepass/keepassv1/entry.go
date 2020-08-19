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

func FormatTime(t time.Time) (formatted string) {
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

func (e *Entry) Output(full bool) (val string) {
	var b strings.Builder
	val = "\n"
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "UUID:\t%s\n", e.entry.UUID)

	fmt.Fprintf(&b, "Creation Time:\t%s\n", FormatTime(e.entry.CreationTime))
	fmt.Fprintf(&b, "Last Modified:\t%s\n", FormatTime(e.entry.LastModificationTime))
	fmt.Fprintf(&b, "Last Accessed:\t%s\n", FormatTime(e.entry.LastAccessTime))

	path, err := e.Path()
	if err != nil {
		path = fmt.Sprintf("<error: %s", err)
	}
	fmt.Fprintf(&b, "Location:\t%s\n", path)
	fmt.Fprintf(&b, "Title:\t%s\n", e.Title())
	fmt.Fprintf(&b, "URL:\t%s\n", e.Get("url").Value)
	fmt.Fprintf(&b, "Username:\t%s\n", e.Get("username").Value)
	password := "[redacted]"
	if full {
		password = e.Password()
	}
	fmt.Fprintf(&b, "Password:\t%s\n", password)
	fmt.Fprintf(&b, "Notes:\n%s\n", e.Get("notes").Value)
	if e.entry.HasAttachment() {
		fmt.Fprintf(&b, "Attachment:\t%s\n", e.Get("attachment").Name)
	}
	return b.String()
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

func (e *Entry) Values() (vals []k.Value) {
	path, _ := e.Path()
	vals = append(vals, k.Value{Name: "location", Value: []byte(path), ReadOnly: true, Searchable: false})
	vals = append(vals, k.Value{Name: fieldTitle, Value: []byte(e.Title()), Searchable: true})
	vals = append(vals, k.Value{Name: fieldUrl, Value: []byte(e.Get("url").Value), Searchable: true})
	vals = append(vals, k.Value{Name: fieldUn, Value: []byte(e.Get("username").Value), Searchable: true})
	vals = append(vals, k.Value{Name: fieldPw, Value: []byte(e.Password()), Searchable: true, Protected: true})
	vals = append(vals, k.Value{Name: fieldNotes, Value: []byte(e.Get("notes").Value), Searchable: true, Type: k.LONGSTRING})
	vals = append(vals, k.Value{Name: fieldAttachment, Value: e.Get("Attachment").Value, Searchable: true, Type: k.BINARY})
	return
}
