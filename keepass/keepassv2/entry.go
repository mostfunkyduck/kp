package keepassv2

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	k "github.com/mostfunkyduck/kp/keepass"
	c "github.com/mostfunkyduck/kp/keepass/common"
	g "github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

type Entry struct {
	c.Entry
	entry *g.Entry
}

func WrapEntry(entry *g.Entry, db k.Database) k.Entry {
	wrapper := &Entry{
		entry: entry,
	}
	wrapper.SetDB(db)
	wrapper.SetDriver(wrapper)
	return wrapper
}

func (e *Entry) Raw() interface{} {
	return e.entry
}

// returns the fully qualified path to the entry, if there's no parent, only the name is returned
func (e *Entry) UUIDString() (string, error) {
	encodedUUID, err := e.entry.UUID.MarshalText()
	if err != nil {
		return "", fmt.Errorf("could not encode UUID: %s", err)
	}
	str, err := base64.StdEncoding.DecodeString(string(encodedUUID))
	if err != nil {
		return "", fmt.Errorf("could not decode b64: %s", err)
	}
	return string(str), nil
}

func (e Entry) Get(field string) k.Value {
	val := e.entry.Get(field)
	if val == nil {
		return k.Value{}
	}

	return k.Value{
		Name:      field,
		Value:     []byte(val.Value.Content),
		Protected: val.Value.Protected.Bool,
	}
}

func (e *Entry) Set(value k.Value) bool {
	for i, each := range e.entry.Values {
		if each.Key == value.Name {
			oldContent := each.Value.Content

			// TODO filter for binaries here, bad shit will happen if you try to attach this way :D
			each.Value.Content = string(value.Value)

			// since we don't get to use pointers, update the slice directly
			e.entry.Values[i] = each

			return (oldContent != string(value.Value))
		}
	}
	// no existing value to update, create it fresh
	e.entry.Values = append(e.entry.Values, g.ValueData{
		Key: value.Name,
		Value: g.V{
			Content:   string(value.Value),
			Protected: w.NewBoolWrapper(value.Protected),
		},
	})
	return true
}

func (e *Entry) LastAccessTime() time.Time {
	return e.entry.Times.LastAccessTime.Time
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.Times.LastAccessTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) LastModificationTime() time.Time {
	return e.entry.Times.LastModificationTime.Time
}
func (e *Entry) SetLastModificationTime(t time.Time) {
	e.entry.Times.LastModificationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) CreationTime() time.Time {
	return e.entry.Times.CreationTime.Time
}

func (e *Entry) SetCreationTime(t time.Time) {
	e.entry.Times.CreationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) Values() (values []k.Value, err error) {
	path, err := e.Path()
	if err != nil {
		return []k.Value{}, fmt.Errorf("could not retrieve entry's path: %s", err)
	}
	values = append(values, k.Value{
		Name:       "location",
		Value:      []byte(path),
		ReadOnly:   true,
		Searchable: false,
	})

	for _, each := range e.entry.Values {
		newValue := k.Value{
			Name:       each.Key,
			Type:       k.STRING,
			Value:      []byte(each.Value.Content),
			Searchable: true, // this may have to change if location is embedded in an entry like it is in v1
			Protected:  each.Value.Protected.Bool,
		}

		// notes are always "long", as are strings where the user already entered a lot of spew
		if len(newValue.Value) > 30 || strings.ToLower(each.Key) == "notes" {
			newValue.Type = k.LONGSTRING
		}

		values = append(values, newValue)
	}

	for _, each := range e.entry.Binaries {
		binary, err := e.DB().Binary(each.Value.ID, each.Name)
		if err != nil {
			return []k.Value{}, fmt.Errorf("could not retrieve binary named '%s' with ID '%d': %s", each.Name, each.Value.ID, err)
		}
		if !binary.Present {
			return []k.Value{}, fmt.Errorf("binary retrieval not implemented, this shouldn't happen on v2, but here we are")
		}
		values = append(values, binary.Value)
	}

	return
}

func (e *Entry) SetPassword(password string) {
	e.Set(k.Value{
		Name:      "Password",
		Value:     []byte(password),
		Type:      k.STRING,
		Protected: true,
	})
}

func (e *Entry) Password() string {
	return e.entry.GetPassword()
}

func (e *Entry) SetTitle(title string) {
	e.Set(k.Value{
		Name:  "Title",
		Type:  k.STRING,
		Value: []byte(title),
	})
}
func (e *Entry) Title() string {
	return e.entry.GetTitle()
}
