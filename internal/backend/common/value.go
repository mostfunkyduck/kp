package common

import (
	"fmt"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"strings"
)

type Value struct {
	value      []byte
	name       string // v1 compatibility - attachments have their own name within entries
	searchable bool   // indicates whether this value should be included in searches
	protected  bool
	readOnly   bool
	valueType  t.ValueType
}

// NewValue initializes a value object
func NewValue(value []byte, name string, searchable bool, protected bool, readOnly bool, valueType t.ValueType) Value {
	return Value{
		value:      value,
		name:       name,
		searchable: searchable,
		protected:  protected,
		readOnly:   readOnly,
		valueType:  valueType,
	}
}

// FormattedValue returns the appropriately formatted value contents, with the `full` argument determining
// whether protected values should be returned in cleartext
func (v Value) FormattedValue(full bool) string {
	if v.Type() == t.BINARY {
		return fmt.Sprintf("binary: %d bytes", len(v.value))
	}

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

func (v Value) Value() []byte {
	return v.value
}
func (v Value) Name() string {
	return v.name
}

func (v Value) Searchable() bool {
	return v.searchable
}

func (v Value) Protected() bool {
	return v.protected
}

func (v Value) ReadOnly() bool {
	return v.readOnly
}

func (v Value) Type() t.ValueType {
	return v.valueType
}
