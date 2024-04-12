package common

import (
	"fmt"
	"strings"

	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EntryValue struct {
	value      []byte
	name       string
	searchable bool // indicates whether this value should be included in searches
	protected  bool
	readOnly   bool
	valueType  t.ValueType
}

type Attachment struct {
	EntryValue
}

// NewValue initializes a value object
func NewValue(value []byte, name string, searchable bool, protected bool, readOnly bool, valueType t.ValueType) EntryValue {
	return EntryValue{
		value:      value,
		name:       name,
		searchable: searchable,
		protected:  protected,
		readOnly:   readOnly,
		valueType:  valueType,
	}
}

func (a Attachment) FormattedValue(full bool) string {
	return fmt.Sprintf("binary: %d bytes", len(a.value))
}

// FormattedValue returns the appropriately formatted value contents, with the `full` argument determining
// whether protected values should be returned in cleartext
func (v EntryValue) FormattedValue(full bool) string {

	if v.Protected() && !full {
		return "[protected]"
	}

	if v.Type() == t.LONGSTRING {
		value := string(v.Value())
		// Long fields are going to need a line break so the first line isn't corrupted
		value = "\n" + value

		// Add indentations for all line breaks to differentiate note lines from field lines
		value = strings.ReplaceAll(value, "\n", "\n>\t")
		return value
	}
	return string(v.Value())
}

func (v EntryValue) Value() []byte {
	return v.value
}

func (v EntryValue) NameTitle() string {
	return cases.Title(language.English, cases.NoLower).String(v.name)
}

func (v EntryValue) Output(showProtected bool) string {
	return fmt.Sprintf("%s:\t%s", v.NameTitle(), v.FormattedValue(showProtected))
}

func (a Attachment) Output(showProtected bool) string {
	return fmt.Sprintf("Attachment:\n\tName:\t%s\n\tSize:\t%s", a.Name(), a.FormattedValue(showProtected))
}

func (v EntryValue) Name() string {
	return v.name
}

func (v EntryValue) Searchable() bool {
	return v.searchable
}

func (v EntryValue) Protected() bool {
	return v.protected
}

func (v EntryValue) ReadOnly() bool {
	return v.readOnly
}

func (v EntryValue) Type() t.ValueType {
	return v.valueType
}
