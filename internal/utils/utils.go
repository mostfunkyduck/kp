package utils

import (
	"fmt"
	"os"
)

// RemoveLockfile removes the lock file on the current savepath of the database
func RemoveLockfile(filePath string) error {
	if filePath != "" {
		if err := os.Remove(filePath + ".lock"); err != nil {
			return fmt.Errorf("could not remove lockfile: %s", err)
		}
	}
	return nil
}
