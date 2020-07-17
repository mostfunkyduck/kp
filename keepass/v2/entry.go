package keepassv2

import (
	k "github.com/mostfunkyduck/kp/keepass"
	g "github.com/tobischo/gokeepasslib/v3"
)
type Entry struct {
	entry *g.Entry
}
func newEntry(entry *g.Entry) k.Entry {
	return Entry{
		entry: entry,
	}
}
