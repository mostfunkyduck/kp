package keepassv1_test

import (
	"testing"

	runner "github.com/mostfunkyduck/kp/internal/backend/tests"
)

func TestRegularPath(t *testing.T) {
	r := createTestResources(t)
	runner.RunTestRegularPath(t, r)
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
