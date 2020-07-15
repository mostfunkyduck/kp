package keepassv1_test

import (
	v1 "github.com/mostfunkyduck/kp/keepass/v1"
	"testing"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func TestTitle(t *testing.T) {
	title := "test"
	e := &keepass.Entry{Title: title}
	wrapper := v1.NewEntry(e)
	wrapperTitle := string(wrapper.Get("title").Value())
	if wrapperTitle != title {
		t.Fatalf("%s != %s", title, wrapperTitle)
	}
}
