package main

// Tests the shell commands
// This will test shell output because there's not really any other way to do it.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

// the Write() method can't store output locally b/c it isn't a pointer target
// this is the workaround
type outputHolder struct {
	output string
}
type FakeWriter struct {
	outputHolder *outputHolder
}

func (f FakeWriter) Write(p []byte) (n int, err error) {
	// output will look a little funny...
	f.outputHolder.output += strings.TrimSpace(strings.ReplaceAll(string(p), "\n", " "))

	return len(p), nil
}

type testResources struct {
	Shell   *ishell.Shell
	Context *ishell.Context
	Group   *keepass.Group
	Path    string
	Db      *keepass.Database
	Entry   *keepass.Entry
	F       FakeWriter
}

func createTestResources(t *testing.T) (r testResources) {
	r.Shell = ishell.New()
	r.Path = "test/test"
	r.Context = &ishell.Context{}
	var err error
	r.Db, err = keepass.New(&keepass.Options{})
	if err != nil {
		t.Fatalf("could not open test db: %s", err)
	}

	r.Group = r.Db.Root().NewSubgroup()
	r.Group.Name = "test"

	r.Entry, err = r.Group.NewEntry()
	if err != nil {
		t.Fatalf("could not create entry: %s", err)
	}
	r.Entry.Title = "test"
	r.Entry.URL = "example.com"
	r.Entry.Username = "username"
	r.Entry.Password = "password"
	r.Entry.Notes = "Notes"

	r.Context.Set("currentLocation", r.Db.Root())
	r.F = FakeWriter{
		outputHolder: &outputHolder{},
	}
	r.Shell.SetOut(r.F)
	return
}

func testEntry(redactedPassword bool, t *testing.T, r testResources) {
	o := r.F.outputHolder.output
	testShowOutput(o, fmt.Sprintf("Location: %s", r.Path), t)
	testShowOutput(o, fmt.Sprintf("Title: %s", r.Entry.Title), t)
	testShowOutput(o, fmt.Sprintf("URL: %s", r.Entry.URL), t)
	testShowOutput(o, fmt.Sprintf("Username: %s", r.Entry.Username), t)
	if redactedPassword {
		testShowOutput(o, "Password: [redacted]", t)
	} else {
		testShowOutput(o, fmt.Sprintf("Password: %s", r.Entry.Password), t)
	}

	testShowOutput(o, fmt.Sprintf("Notes: %s", r.Entry.Notes), t)

	if r.Entry.HasAttachment() {
		testShowOutput(o, fmt.Sprintf("Attachment: %s", r.Entry.Attachment.Name), t)
	}
}

func testShowOutput(output string, substr string, t *testing.T) {
	if !strings.Contains(output, substr) {
		t.Errorf("output [%s] does not contain expected string [%s]", output, substr)
	}
}

func TestShowNoArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}
	cmd := ishell.Cmd{
		Help: "test string",
	}
	r.Context.Cmd = cmd
	Show(r.Shell)(r.Context)
	expected := "syntax: " + r.Context.Cmd.Help
	if r.F.outputHolder.output != expected {
		t.Fatalf("output was incorrect: %s != %s", r.F.outputHolder.output, expected)
	}
}

func TestShowValidArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Path}
	Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowAttachment(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Path}
	r.Entry.Attachment.Name = "asdf"

	Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowFullMode(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{"-f", r.Path}
	Show(r.Shell)(r.Context)
	testEntry(false, t, r)
}

func TestCdToGroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{
		r.Group.Name,
	}

	r.Shell.Set("currentLocation", r.Group)

	Cd(r.Shell)(r.Context)

	l := r.Shell.Get("currentLocation").(*keepass.Group)
	if l.ID != r.Db.Root().ID {
		t.Fatalf("new location was not the one specified: %d != %d", l.ID, r.Group.ID)
	}
}

func TestCdToRoot(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}

	r.Shell.Set("currentLocation", r.Group)

	Cd(r.Shell)(r.Context)

	l := r.Shell.Get("currentLocation").(*keepass.Group)
	if l.ID != r.Db.Root().ID {
		t.Fatalf("new location was not the one specified: %d != %d", l.ID, r.Group.ID)
	}
}

func TestCdToSubgroup(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{
		r.Group.Name,
	}

	r.Shell.Set("currentLocation", r.Db.Root())

	Cd(r.Shell)(r.Context)

	l := r.Shell.Get("currentLocation").(*keepass.Group)
	if l.ID != r.Group.ID {
		t.Fatalf("new location was not the one specified: %d != %d", l.ID, r.Group.ID)
	}
}

func TestLsNoArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}
	Ls(r.Shell)(r.Context)
	if r.F.outputHolder.output != "test/" {
		t.Fatalf("output did not contain group: %s != %s", r.F.outputHolder.output, "test/")
	}
}

func TestLsArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Path}
	Ls(r.Shell)(r.Context)
	if r.F.outputHolder.output != "0: test" {
		t.Fatalf("output did not contain group: %s != %s", r.F.outputHolder.output, "test/")
	}
}
