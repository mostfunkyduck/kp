package keepassv1_test

import (
	"testing"

	v1 "github.com/mostfunkyduck/kp/keepass/keepassv1"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func TestTitle(t *testing.T) {
	title := "test"
	e := &keepass.Entry{Title: title}
	wrapper := v1.NewEntry(e)
	wrapperTitle := wrapper.Get("title").Value.(string)
	if wrapperTitle != title {
		t.Fatalf("%s != %s", title, wrapperTitle)
	}
}
