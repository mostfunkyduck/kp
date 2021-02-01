package main_test

// Scaffolding for running shell command tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	k "github.com/mostfunkyduck/kp/keepass"
	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	v2 "github.com/mostfunkyduck/kp/keepass/keepassv2"
	keepass2 "github.com/tobischo/gokeepasslib/v3"
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
	Shell    *ishell.Shell
	Context  *ishell.Context
	Group    k.Group
	Path     string
	Db       k.Database
	Entry    k.Entry
	F        FakeWriter
	Readline *readline.Instance
}

func initDBv1() (k.Database, error) {
	db, err := keepass.New(&keepass.Options{KeyRounds: 1})
	if err != nil {
		return nil, fmt.Errorf("could not open test db: %s", err)
	}

	return v1.NewDatabase(db, ""), nil
}

func initDBv2() (k.Database, error) {
	db := keepass2.NewDatabase()
	dbWrapper := v2.NewDatabase(db, "", k.Options{})
	return dbWrapper, nil
}

func createTestResources(t *testing.T) (r testResources) {
	var err error
	r.Readline, err = readline.New("")
	if err != nil {
		t.Fatal(err)
	}
	r.Shell = ishell.NewWithReadline(r.Readline)
	r.Path = "test/test"
	r.Context = &ishell.Context{}
	version := os.Getenv("KPVERSION")
	if version == "1" {
		r.Db, err = initDBv1()
	} else if version == "2" {
		r.Db, err = initDBv2()
	} else {
		t.Fatalf("KPVERSION environment variable invalid (value: '%s'), rerun with it as either '1' or '2'", version)
	}
	if err != nil {
		t.Fatal(err)
	}
	r.Shell.Set("db", r.Db)
	r.Group, _ = r.Db.Root().NewSubgroup("test")

	r.Entry, err = r.Group.NewEntry("test")
	if err != nil {
		t.Fatalf("could not create entry: %s", err)
	}
	settings := map[string]string{
		"Title":    "test",
		"URL":      "example.com",
		"UserName": "username",
		"Password": "password",
		"Notes":    "notes",
	}
	for key, v := range settings {
		val := k.Value{
			Name:  key,
			Value: []byte(v),
		}
		r.Entry.Set(val)
	}

	r.F = FakeWriter{
		outputHolder: &outputHolder{},
	}
	r.Shell.SetOut(r.F)
	return
}

func testEntry(redactedPassword bool, t *testing.T, r testResources) {
	o := r.F.outputHolder.output
	path, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	testShowOutput(o, fmt.Sprintf("Location:\t%s", path), t)
	testShowOutput(o, fmt.Sprintf("Title:\t%s", r.Entry.Title()), t)
	testShowOutput(o, fmt.Sprintf("URL:\t%s", r.Entry.Get("URL").Value), t)
	testShowOutput(o, fmt.Sprintf("Username:\t%s", r.Entry.Get("UserName").Value), t)
	if redactedPassword {
		testShowOutput(o, "Password:\t[redacted]", t)
	} else {
		testShowOutput(o, fmt.Sprintf("Password:\t%s", r.Entry.Password()), t)
	}

	// format the notes to match how the entry will format long strings for output, which is not how they're stored internally
	// This is ridiculously annoying to test properly, pushing it off for now, will test manually
	//testShowOutput(o, fmt.Sprintf("Notes:\t\n>\t%s", strings.ReplaceAll(string(r.Entry.Get("notes").Value), "\n", "\n>\t")), t)

	att := r.Entry.Get("attachment")
	if len(att.Value) != 0 || att.Name != "" {
		testShowOutput(o, fmt.Sprintf("Attachment:\t%s", att.Name), t)
	}
}
