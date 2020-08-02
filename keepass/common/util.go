package common

import (
	k "github.com/mostfunkyduck/kp/keepass"
)

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
