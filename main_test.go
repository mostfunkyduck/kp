package main_test

// Scaffolding for running shell command tests

import (
	"fmt"
	"strings"
	"testing"

	k "github.com/mostfunkyduck/kp/keepass"
	v1 "github.com/mostfunkyduck/kp/keepass/v1"
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
	Group   k.Group
	Path    string
	Db      k.Database
	Entry   k.Entry
	F       FakeWriter
}

func createTestResources(t *testing.T) (r testResources) {
	r.Shell = ishell.New()
	r.Path = "test/test"
	r.Context = &ishell.Context{}
	var err error
	db, err := keepass.New(&keepass.Options{})
	if err != nil {
		t.Fatalf("could not open test db: %s", err)
	}

	r.Db = v1.NewDatabase(db, "")
	r.Shell.Set("db", r.Db)
	r.Group = r.Db.Root().NewSubgroup("test")

	r.Entry, err = r.Group.NewEntry()
	if err != nil {
		t.Fatalf("could not create entry: %s", err)
	}
	settings := map[string]string {
		"title": "test",
		"url": "example.com",
		"username": "username",
		"password": "password",
		"notes": "notes",
	}
	for key, v := range settings {
		val := k.Value{
			Name: key,
			Value: v,
		}
		r.Entry.Set(key, val)
	}

	r.F = FakeWriter{
		outputHolder: &outputHolder{},
	}
	r.Shell.SetOut(r.F)
	return
}

func testEntry(redactedPassword bool, t *testing.T, r testResources) {
	o := r.F.outputHolder.output
	testShowOutput(o, fmt.Sprintf("Location:\t%s", r.Entry.Pwd()), t)
	testShowOutput(o, fmt.Sprintf("Title:\t%s", r.Entry.Get("title").Value), t)
	testShowOutput(o, fmt.Sprintf("URL:\t%s", r.Entry.Get("url").Value), t)
	testShowOutput(o, fmt.Sprintf("Username:\t%s", r.Entry.Get("username").Value), t)
	if redactedPassword {
		testShowOutput(o, "Password:\t[redacted]", t)
	} else {
		testShowOutput(o, fmt.Sprintf("Password:\t%s", r.Entry.Get("password").Value), t)
	}

	testShowOutput(o, fmt.Sprintf("Notes: %s", r.Entry.Get("notes").Value), t)

	att := r.Entry.Get("attachment")
	if att != (k.Value{}) {
		testShowOutput(o, fmt.Sprintf("Attachment:\t%s", att.Name), t)
	}
}
