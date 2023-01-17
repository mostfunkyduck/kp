package commands_test

// Scaffolding for running shell command tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/abiosoft/readline"
	"github.com/mostfunkyduck/ishell"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	v2 "github.com/mostfunkyduck/kp/internal/backend/keepassv2"
	"github.com/mostfunkyduck/kp/internal/backend/types"
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
	Group    types.Group
	Path     string
	Db       types.Database
	Entry    types.Entry
	F        FakeWriter
	Readline *readline.Instance
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

	tmpFile, err := os.CreateTemp("", "kp_unit_tests")
	if err != nil {
		t.Fatal(err)
	}

	// this will allow the init code to create the db from scratch
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	// this will remove it afterwards, which will break everything if the tests try to save
	// but save tests can run their own init
	defer os.Remove(tmpFile.Name())
	opts := types.Options{
		DBPath:    tmpFile.Name(),
		KeyRounds: 1,
	}
	if version == "1" {
		r.Db = &v1.Database{}
	} else if version == "2" {
		r.Db = &v2.Database{}
	} else {
		t.Fatalf("KPVERSION environment variable invalid (value: '%s'), rerun with it as either '1' or '2'", version)
	}

	if err := r.Db.Init(opts); err != nil {
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
		val := c.NewValue(
			[]byte(v),
			key,
			false, false, false,
			types.STRING,
		)

		r.Entry.Set(val)
	}

	r.F = FakeWriter{
		outputHolder: &outputHolder{},
	}
	r.Shell.SetOut(r.F)
	return
}

func testEntry(full bool, t *testing.T, r testResources) {
	o := r.F.outputHolder.output
	path, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	testShowOutput(o, fmt.Sprintf("Location:\t%s", path), t)
	testShowOutput(o, fmt.Sprintf("Title:\t%s", r.Entry.Title()), t)
	testShowOutput(o, fmt.Sprintf("URL:\t%s", r.Entry.Get("URL").Value()), t)
	// compensating for v1 and v2 formatting differently
	unFieldName := "Username"
	if os.Getenv("KPVERSION") == "2" {
		unFieldName = "UserName"
	}
	testShowOutput(o, fmt.Sprintf("%s:\t%s", unFieldName, r.Entry.Username()), t)
	if full {
		testShowOutput(o, "Password:\t[protected]", t)
	} else {
		testShowOutput(o, fmt.Sprintf("Password:\t%s", r.Entry.Password()), t)
	}

	// format the notes to match how the entry will format long strings for output, which is not how they're stored internally
	// This is ridiculously annoying to test properly, pushing it off for now, will test manually
	//testShowOutput(o, fmt.Sprintf("Notes:\t\n>\t%s", strings.ReplaceAll(string(r.Entry.Get("notes").Value), "\n", "\n>\t")), t)

	att := r.Entry.Get("attachment")
	if att != nil {
		testShowOutput(o, fmt.Sprintf("Attachment:\t%s", att.Name()), t)
	}
}
