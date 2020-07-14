package keepassv1

import (
	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Entry struct {
	entry *keepass.Entry
}

func NewEntry(entry *keepass.Entry) k.Entry {
	return Entry{
		entry: entry,
	}
}

func (e Entry) Title() string {
	return e.entry.Title
}

