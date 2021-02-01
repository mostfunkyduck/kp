package keepassv2_test

import (
	"reflect"
	"strings"
	"testing"

	k "github.com/mostfunkyduck/kp/keepass"
	main "github.com/mostfunkyduck/kp/keepass/keepassv2"
	runner "github.com/mostfunkyduck/kp/keepass/tests"
	g "github.com/tobischo/gokeepasslib/v3"
)

func TestNoParent(t *testing.T) {
	r := runner.Resources{}
	r.Db = main.NewDatabase(g.NewDatabase(), "/dev/null", k.Options{})
	newEnt := g.NewEntry()
	r.Entry = main.WrapEntry(&newEnt, r.Db)

	runner.RunTestNoParent(t, r)
}

func TestNewEntry(t *testing.T) {
	r := createTestResources(t)
	newEnt, _ := r.Group.NewEntry("newentry")
	// loop through the default values and make sure that each of the expected ones are in there
	// this is indicated by flipping the bool in the expectedFields map to 'true', we will
	// then test that each field was set to 'true' during the loop
	expectedFields := map[string]bool{
		"notes":    false,
		"username": false,
		"password": false,
		"title":    false,
		"url":      false,
	}
	values, err := newEnt.Values()
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range values {
		lcName := strings.ToLower(val.Name)
		if _, present := expectedFields[lcName]; present {
			expectedFields[lcName] = true
		}
	}

	for k, v := range expectedFields {
		if !v {
			t.Fatalf("field [%s] was not present in new entry\n", k)
		}
	}
}
func TestRegularPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestRegularPath(t, r)
}

func TestEntryGetSet(t *testing.T) {
	r := createTestResources(t)
	value := k.Value{
		Name:  "TestEntrySetGet",
		Value: []byte("test value"),
	}

	retVal := r.BlankEntry.Get(value.Name)
	blankValue := k.Value{}
	if !reflect.DeepEqual(retVal, blankValue) {
		t.Fatalf("[%v] != [%v]", retVal, blankValue)
	}
	if !r.BlankEntry.Set(value) {
		t.Fatalf("could not set value")
	}

	entryValue := string(r.BlankEntry.Get(value.Name).Value)
	if entryValue != string(value.Value) {
		t.Fatalf("[%s] != [%s], %v", entryValue, value.Name, value)
	}

	secondValue := "asldkfj"
	value.Value = []byte(secondValue)
	if !r.BlankEntry.Set(value) {
		t.Fatalf("could not overwrite value: %v", value)
	}

	entryValue = string(r.BlankEntry.Get(value.Name).Value)
	if entryValue != secondValue {
		t.Fatalf("[%s] != [%s] %v", entryValue, secondValue, value)
	}
}

func TestEntryTimeFuncs(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestEntryTimeFuncs(t, r)
}

func TestEntryPasswordTitleFuncs(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestEntryPasswordTitleFuncs(t, r)
}

func TestOutput(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestOutput(t, r.Entry)
}
