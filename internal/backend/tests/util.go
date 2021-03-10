package tests

import (
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

type Resources struct {
	Db    t.Database
	Entry t.Entry
	Group t.Group
	// BlankEntry and BlankGroup are empty resources for testing freshly
	// allocated structs
	BlankEntry t.Entry
	BlankGroup t.Group
}
