package main_test

import (
	"strings"
	"testing"
	main "github.com/mostfunkyduck/kp"
	k "github.com/mostfunkyduck/kp/keepass"
	"github.com/abiosoft/ishell"
)
func testShowOutput(output string, substr string, t *testing.T) {
	if !strings.Contains(output, substr) {
		t.Errorf("output [%s] does not contain expected string [%s]", output, substr)
	}
}

// 'show' with no arguments should error out
func TestShowNoArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{}
	cmd := ishell.Cmd{
		Help: "test string",
	}
	r.Context.Cmd = cmd
	main.Show(r.Shell)(r.Context)
	expected := "syntax: " + r.Context.Cmd.Help
	if r.F.outputHolder.output != expected {
		t.Fatalf("output was incorrect: %s != %s", r.F.outputHolder.output, expected)
	}
}

func TestShowValidArgs(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Entry.Pwd()}
	main.Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowAttachment(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Path}
	att := k.Value{
		Name: "asdf",
		Value: []byte("yaakov is cool"),
	}
	r.Entry.Set("attachment", att)

	main.Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowFullMode(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{"-f", r.Path}
	main.Show(r.Shell)(r.Context)
	testEntry(false, t, r)
}


