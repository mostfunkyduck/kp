package keepassv1_test

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	v1 "github.com/mostfunkyduck/kp/internal/backend/keepassv1"
	"github.com/mostfunkyduck/kp/internal/backend/types"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func TestTitle(t *testing.T) {
	title := "test"
	e := &keepass.Entry{Title: title}
	wrapper := v1.WrapEntry(e, &v1.Database{})
	wrapperTitle := wrapper.Title()
	if wrapperTitle != title {
		t.Fatalf("%s != %s", title, wrapperTitle)
	}
}

func TestEntrySearch(t *testing.T) {
	title := "TestEntrySearch"
	tmpfile, err := ioutil.TempFile("", "kp_unit_tests")
	if err != nil {
		t.Fatalf("could not create temp file for DB: %s", tmpfile.Name())
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())
	defer os.Remove(tmpfile.Name())

	dbWrapper := v1.Database{}
	dbOptions := types.Options{
		DBPath:    tmpfile.Name(),
		KeyRounds: 1,
	}
	if err := dbWrapper.Init(dbOptions); err != nil {
		t.Fatal(err)
	}
	sg, err := dbWrapper.Root().NewSubgroup("DOESN'T MATCH")
	if err != nil {
		t.Fatal(err)
	}

	wrapper, err := sg.NewEntry(title)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := wrapper.Search(regexp.MustCompile("TestEntry.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("%v", paths)
	}

	path, err := wrapper.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if paths[0] != path {
		t.Fatalf("[%s] != [%s]", paths[0], path)
	}
}
