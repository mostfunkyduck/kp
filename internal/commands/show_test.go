package commands_test

import (
	"strings"
	"testing"

	"github.com/mostfunkyduck/ishell"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	"github.com/mostfunkyduck/kp/internal/backend/types"
	main "github.com/mostfunkyduck/kp/internal/commands"
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
	path, err := r.Entry.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Context.Args = []string{path}
	main.Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowAttachment(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{r.Path}
	att := c.NewValue(
		[]byte("yaakov is cool"),
		"asdf",
		false, false, false,
		types.BINARY,
	)

	r.Entry.Set(att)

	main.Show(r.Shell)(r.Context)

	testEntry(true, t, r)
}

func TestShowFullMode(t *testing.T) {
	r := createTestResources(t)
	r.Context.Args = []string{"-f", r.Path}
	r.Context.Flags = []string{"-f"}
	main.Show(r.Shell)(r.Context)
	testEntry(false, t, r)
}
