package keepassv1_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
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
	db, err := keepass.New(&keepass.Options{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	dbWrapper := v1.NewDatabase(db, "/dev/null")
	sg, err := dbWrapper.Root().NewSubgroup("DOESN'T MATCH")
	if err != nil {
		t.Fatalf(err.Error())
	}

	wrapper, err := sg.NewEntry(title)
	if err != nil {
		t.Fatalf(err.Error())
	}
	paths := wrapper.Search(regexp.MustCompile("TestEntry.*"))
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

func TestFormatTime(t *testing.T) {
	then := time.Now().Add(time.Duration(-1) * time.Hour * 24)
	str := v1.FormatTime(then)
	expected := "1 days ago"
	if !strings.Contains(str, expected) {
		t.Fatalf("[%s] doesn't contain [%s]", str, expected)
	}

	then = time.Now().Add(time.Duration(-1) * time.Hour * 24 * 35)
	str = v1.FormatTime(then)
	expected = "about 1 months ago"
	if !strings.Contains(str, expected) {
		t.Fatalf("[%s] doesn't contain [%s]", str, expected)
	}

	then = time.Now().Add(time.Duration(-1) * time.Hour * 24 * 365)
	str = v1.FormatTime(then)
	expected = "about 1 years ago"
	if !strings.Contains(str, expected) {
		t.Fatalf("[%s] doesn't contain [%s]", str, expected)
	}


	then = time.Now()
	str = v1.FormatTime(then)
	expected = "less than a second ago"
	if !strings.Contains(str, expected) {
		t.Fatalf("[%s] doesn't contain [%s]", str, expected)
	}
}
