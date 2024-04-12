package commands_test

import (
	"strings"
	"testing"

	main "github.com/mostfunkyduck/kp/internal/commands"
)

// Tests ls within a group that contains a subgroup and an entry
func TestLsNoArgsFromGroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}
	_, _ = r.Group.NewSubgroup("test")
	r.Db.SetCurrentLocation(r.Group)
	main.Ls(r.Shell)(r.Context)
	expected := "=== Groups ===\ntest/\n\n=== Entries ===\n0: test\n"
	if !strings.Contains(r.F.outputHolder.output, expected) {
		t.Fatalf("[%s] does not contain  [%s]", r.F.outputHolder.output, expected)
	}
}

func TestLsEntryFromRoot(t *testing.T) {
	r := createTestResources(t)
	path, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{path}
	r.Db.SetCurrentLocation(r.Db.Root())
	main.Ls(r.Shell)(r.Context)
	if r.F.outputHolder.output != "test\n" {
		t.Fatalf("[%s] does not contain  [%s]", r.F.outputHolder.output, "test")
	}
}
