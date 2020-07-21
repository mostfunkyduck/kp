package keepassv2

import (
	k "github.com/mostfunkyduck/kp/keepass"
)

// FIXME this belongs in the parent directory, along with a bunch of other stuff
func CompareUUIDs(me k.UUIDer, them k.UUIDer) (bool, error) {
	myUUID, err := me.UUIDString()
	if err != nil {
		return false, err
	}

	theirUUID, err := them.UUIDString()
	if err != nil {
		return false, err
	}

	return theirUUID == myUUID, nil
}
