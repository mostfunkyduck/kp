package common

import (
	"fmt"

	k "github.com/mostfunkyduck/kp/keepass"
)

func findPathToGroup(source k.Group, target k.Group) (rv []k.Group, err error) {
	// this library doesn't appear to support child->parent links, so we have to find the needful ourselves
	for _, group := range source.Groups() {
		same, err := CompareUUIDs(group, target)
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not compare UUIDS: %s", err)
		}

		if same {
			return []k.Group{source}, nil
		}

		pathGroups, err := findPathToGroup(group, target)
		if err != nil {
			return []k.Group{}, fmt.Errorf("could not find path from group '%s' to group '%s': %s", group.Name(), target.Name(), err)
		}

		if len(pathGroups) != 0 {
			return append([]k.Group{source}, pathGroups...), nil
		}
	}
	return []k.Group{}, nil
}
