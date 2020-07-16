package main_test

import (
	main "github.com/mostfunkyduck/kp"
	"testing"
)

// prepares stdin to fill out a new entry with default values and decline to save
func fillOutEntry(r testResources) error {
	for _, each := range []string{"\n", "\n", "\n", "\n", "\n", "N", "n"} {
		if _, err := r.Readline.WriteStdin([]byte(each)); err != nil {
			return err
		}
	}
	return nil
}

func TestNewEntry(t *testing.T) {
	r := createTestResources(t)
	entryName := "asdfsadf"
	r.Db.SetCurrentLocation(r.Group)
	originalEntriesLen := len(r.Group.Entries())
	r.Context.Args = []string{
		entryName,
	}

	if err := fillOutEntry(r); err != nil {
		t.Fatalf(err.Error())
	}
	main.NewEntry(r.Shell)(r.Context)
	output := r.F.outputHolder.output
	entries := r.Group.Entries()
	if len(entries) != originalEntriesLen+1 {
		t.Fatalf("wrong number of entries after initial entry creation: [%d] != [%d] (%s)", len(entries), originalEntriesLen+1, output)
	}

	expectedPath := r.Group.Pwd() + entryName
	// assuming that ordering is deterministic, if it isn't then this test will randomly fail
	if entries[1].Pwd() != expectedPath {
		t.Fatalf("[%s] != [%s] (%s)", entries[1].Pwd(), expectedPath, output)
	}
}

func TestNewAtRoot(t *testing.T) {
	r := createTestResources(t)
	entryName := "asdlfkjsdflkjasdflkj"
	r.Context.Args = []string{
		"/" + entryName,
	}

	main.NewEntry(r.Shell)(r.Context)
	if len(r.Db.Root().Entries()) != 0 {
		t.Fatalf("entry created at root, [%d] != [%d]", len(r.Db.Root().Entries()), 0)
	}
}

func TestDuplicateEntry(t *testing.T) {
	r := createTestResources(t)
	entryName := "taslkfdj"
	r.Context.Args = []string{
		entryName,
	}

	if err := fillOutEntry(r); err != nil {
		t.Fatalf(err.Error())
	}
	main.NewEntry(r.Shell)(r.Context)
	originalEntriesLen := len(r.Db.CurrentLocation().Entries())
	main.NewEntry(r.Shell)(r.Context)

	if len(r.Db.CurrentLocation().Entries()) != originalEntriesLen {
		t.Fatalf("created duplicate entry: [%d] != [%d]", len(r.Db.CurrentLocation().Entries()), originalEntriesLen)
	}
}
