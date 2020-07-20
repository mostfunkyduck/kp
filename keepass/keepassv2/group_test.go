package keepassv2_test

import (
	"testing"
)

func TestNestedSubGroupPath(t *testing.T) {
	r := createTestResources(t)
	sgName := "blipblip"
	sg, err := r.Group.NewSubgroup(sgName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	path, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "/" + r.Group.Name() + "/" + sgName
	if path != expected {
		t.Fatalf("[%s] != [%s]", path, expected)
	}
}

func TestDoubleNestedGroupPath(t *testing.T) {
	r := createTestResources(t)
	sgName := "blipblip"
	sg, err := r.Group.NewSubgroup(sgName)
	if err != nil {
		t.Fatalf(err.Error())
	}

	sg1, err := sg.NewSubgroup(sgName + "1")
	if err != nil {
		t.Fatalf(err.Error())
	}

	sgPath, err := sg.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	sg1Path, err := sg1.Path()
	if err != nil {
		t.Fatalf(err.Error())
	}

	sgExpected := "/" + r.Group.Name() + "/" + sgName
	if sgPath != sgExpected {
		t.Fatalf("[%s] != [%s]", sgPath, sgExpected)
	}

	sg1Expected := sgExpected + "/" + sgName + "1"
	if sg1Path != sg1Expected {
		t.Fatalf("[%s] != [%s]", sgPath, sgExpected)
	}
}
