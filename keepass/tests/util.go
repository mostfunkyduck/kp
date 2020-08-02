package tests

import (
	k "github.com/mostfunkyduck/kp/keepass"
)

type Resources struct {
	Db    k.Database
	Entry k.Entry
	Group k.Group
	// BlankEntry and BlankGroup are empty resources for testing freshly
	// allocated structs
	BlankEntry k.Entry
	BlankGroup k.Group
}
