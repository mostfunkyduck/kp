package commands_test

import (
	"fmt"
	"testing"

	"github.com/mostfunkyduck/kp/internal/backend/types"
	main "github.com/mostfunkyduck/kp/internal/commands"
)

// prepares stdin to fill out a new entry with default values and decline to save
var entryValues = []string{
	"first\n",
	"second\n",
	"third\n",
	"fourth\n", "fourth\n", // password confirmation
	"\n", // notes open in editor, needs manual verification
}

func fillOutEntry(r testResources) error {
	allValues := append(entryValues, []string{"N", "n"}...)
	for _, each := range allValues {
		if _, err := r.Readline.WriteStdin([]byte(each)); err != nil {
			return err
		}
	}
	return nil
}

// verifyDefaultEntry goes through each of the default v1 values
// and test if they show up as expected in the entry passed in
func verifyDefaultEntry(e types.Entry) error {
	// mild hack, but these are formatted in line with what v2 uses
	// v1 is good enough to do a case insensitive match
	// this could be improved in the future with calls to the utility functions, but works for now
	values := map[string]string{
		"Title":    "first",
		"URL":      "second",
		"UserName": "third",
		"Password": "fourth",
		"Notes":    "",
	}

	for k, v := range values {
		val, _ := e.Get(k)

		if string(val.Value()) != v {
			return fmt.Errorf("%s != %s", v, val)
		}
	}

	return nil
}
func TestNewEntry(t *testing.T) {
	r := createTestResources(t)
	r.Db.SetCurrentLocation(r.Group)
	originalEntriesLen := len(r.Group.Entries())
	r.Context.Args = []string{
		// will be overwritten by fillOutEntry
		"replaceme",
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

	expectedPath, err := r.Group.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// the fillOutEntry form replaced the default title name with this one
	expectedPath += "first"
	// assuming that ordering is deterministic, if it isn't then this test will randomly fail
	entryPath, err := entries[1].Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if entryPath != expectedPath {
		t.Fatalf("[%s] != [%s] (%s)", entryPath, expectedPath, output)
	}
	if err := verifyDefaultEntry(entries[1]); err != nil {
		t.Fatalf(err.Error())
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
