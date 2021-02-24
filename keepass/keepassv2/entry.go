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

// field name constants
const (
	fieldUn    = "UserName"
	fieldPw    = "Password"
	fieldUrl   = "URL"
	fieldNotes = "Notes"
	fieldTitle = "Title"
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
	values, err := e.Values()
	if err != nil {
		// swallowing
		return nil
	}

	for _, value := range values {
		if value.Name() == field {
			return value
		}
	}
	return nil
}

func (e *Entry) Set(value k.Value) bool {
	for i, each := range e.entry.Values {
		if each.Key == value.Name() {
			oldContent := each.Value.Content

			// TODO filter for binaries here, bad shit will happen if you try to attach this way :D
			each.Value.Content = string(value.Value())

			// since we don't get to use pointers, update the slice directly
			e.entry.Values[i] = each

			return (oldContent != string(value.Value()))
		}
	}
	// no existing value to update, create it fresh
	e.entry.Values = append(e.entry.Values, g.ValueData{
		Key: value.Name(),
		Value: g.V{
			Content:   string(value.Value()),
			Protected: w.NewBoolWrapper(value.Protected()),
		},
	})
	return true
}

func (e *Entry) LastAccessTime() time.Time {
	if e.entry.Times.LastAccessTime == nil {
		return time.Time{}
	}
	return e.entry.Times.LastAccessTime.Time
}

func (e *Entry) SetLastAccessTime(t time.Time) {
	e.entry.Times.LastAccessTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) LastModificationTime() time.Time {
	if e.entry.Times.LastModificationTime == nil {
		return time.Time{}
	}
	return e.entry.Times.LastModificationTime.Time
}
func (e *Entry) SetLastModificationTime(t time.Time) {
	e.entry.Times.LastModificationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) CreationTime() time.Time {
	if e.entry.Times.CreationTime == nil {
		return time.Time{}
	}
	return e.entry.Times.CreationTime.Time
}

func (e *Entry) SetCreationTime(t time.Time) {
	e.entry.Times.CreationTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) ExpiredTime() time.Time {
	if e.entry.Times.ExpiryTime == nil {
		return time.Time{}
	}
	return e.entry.Times.ExpiryTime.Time
}

func (e *Entry) SetExpiredTime(t time.Time) {
	e.entry.Times.ExpiryTime = &w.TimeWrapper{Time: t}
}

func (e *Entry) Values() (values []k.Value, err error) {
	// we need to arrange this with the regular, "default" values that appear in v1 coming first
	// to preserve UX and predictability of where the fields appear

	// NOTE: the capitalization/formatting here will be how the default value for the field is rendered
	// default values will be set by comparing the actual values in the underlying entry to this list using strings.EqualFold
	// the code uses the existing formatting if it exists, otherwise it will pull from here
	orderedDefaultValues := []string{fieldTitle, fieldUrl, fieldUn, fieldPw, fieldNotes}

	defaultValues := map[string]k.Value{}
	for _, each := range e.entry.Values {
		valueType := k.STRING

		// notes are always "long", as are strings where the user already entered a lot of spew
		if len(each.Value.Content) > 30 || strings.ToLower(each.Key) == "notes" {
			valueType = k.LONGSTRING
		}

		// build the Value object that will wrap this actual value
		newValue := c.NewValue(
			[]byte(each.Value.Content),
			each.Key,
			true, // this may have to change if location is embedded in an entry like it is in v1
			each.Value.Protected.Bool, false,
			valueType,
		)
		defaultValue := false
		for _, val := range orderedDefaultValues {
			if strings.EqualFold(each.Key, val) {
				defaultValue = true
				defaultValues[val] = newValue
			}
		}
		if !defaultValue {
			values = append(values, newValue)
		}
	}

	// prepend the non-default values with the defaults, in expected order
	defaultValueObjects := []k.Value{}
	for _, val := range orderedDefaultValues {
		valObject := defaultValues[val]
		if valObject == nil {
			protected := strings.EqualFold(val, "password")
			valObject = c.NewValue(
				[]byte(""),
				val,
				true,
				protected,
				false,
				k.STRING,
			)
		}
		defaultValueObjects = append(defaultValueObjects, valObject)
	}
	values = append(defaultValueObjects, values...)

	// Prepend everything with the location
	path, err := e.Path()
	if err != nil {
		return []k.Value{}, fmt.Errorf("could not retrieve entry's path: %s", err)
	}

	values = append([]k.Value{
		c.NewValue(
			[]byte(path),
			"location",
			false, false, true,
			k.STRING,
		),
	}, values...)

	// now append entries for the binaries
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
	e.Set(c.NewValue(
		[]byte(password),
		"Password",
		false, true, false,
		k.STRING,
	))
}

func (e *Entry) Password() string {
	return e.entry.GetPassword()
}

func (e *Entry) SetTitle(title string) {
	e.Set(c.NewValue(
		[]byte(title),
		"Title",
		true, false, false,
		k.STRING,
	))
}
func (e *Entry) Title() string {
	return e.entry.GetTitle()
}

func (e *Entry) Username() string {
	return e.Get(fieldUn).FormattedValue(true)
}

func (e *Entry) SetUsername(name string) {
	e.Set(c.NewValue(
		[]byte(name),
		fieldUn,
		true, false, false,
		k.STRING,
	))
}
