package keepassv2

import (
	"encoding/base64"
	"fmt"
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
	wrapper.SetEntry(wrapper)
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
		Name:  field,
		Value: val.Value.Content,
	}
}

func (e *Entry) Set(value k.Value) bool {
	for i, each := range e.entry.Values {
		if each.Key == value.Name {
			oldContent := each.Value.Content
			oldProtected := each.Value.Protected

			// TODO filter for binaries here, bad shit will happen if you try to attach this way :D
			each.Value.Content = value.Value.(string)
			each.Value.Protected = w.NewBoolWrapper(value.Protected)

			// since we don't get to use pointers, update the slice directly
			e.entry.Values[i] = each

			return (oldContent != value.Value) || (oldProtected.Bool != value.Protected)
		}
	}
	// no existing value to update, create it fresh
	e.entry.Values = append(e.entry.Values, g.ValueData{
		Key: value.Name,
		Value: g.V{
			Content:   value.Value.(string),
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

func (e *Entry) Values() (values []k.Value) {
	for _, each := range e.entry.Values {
		newValue := k.Value{
			Name:      each.Key,
			Value:     each.Value.Content,
			Protected: each.Value.Protected.Bool,
		}
		values = append(values, newValue)
	}
	return
}

func (e *Entry) SetPassword(password string) {
	e.Set(k.Value{
		Name:  "Password",
		Value: password,
	})
}

func (e *Entry) Password() string {
	return e.entry.GetPassword()
}

func (e *Entry) SetTitle(title string) {
	e.Set(k.Value{
		Name:  "Title",
		Value: title,
	})
}
func (e *Entry) Title() string {
	return e.entry.GetTitle()
}
