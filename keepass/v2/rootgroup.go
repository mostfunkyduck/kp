package keepassv2

// Unlike the keepass 1 library, this library doesn't represent Root as a group
// which means that we have to dress up its 'RootData' object as a Group object

import (
	g "github.com/tobischo/gokeepasslib/v3"
)

type RootGroup struct {
	Group
	root *g.RootData
}
