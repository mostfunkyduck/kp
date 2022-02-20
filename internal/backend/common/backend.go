package common

import (
	"errors"
	"fmt"
	"os"
)

type Backend struct {
	hash     string
	filename string
}

func InitBackend(filename string) (*Backend, error) {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		// file doesn't exist, so return an empty hash, let the upstream handle creating
		return &Backend{
			filename: filename,
			hash:     "",
		}, nil
	}
	hash, err := GenerateFileHash(filename)
	if err != nil {
		return &Backend{}, fmt.Errorf("could not generate backend hash: %s", err)
	}
	return &Backend{
		filename: filename,
		hash:     hash,
	}, nil
}

// IsModified determines whether or not the underlying storage has been modified since the utility was opened, indicating that something will get stomped
func (b Backend) IsModified() (bool, error) {
	if b.Hash() == "" {
		// this is a new file, consider it unmodified
		return false, nil
	}
	hash, err := GenerateFileHash(b.Filename())
	if err != nil {
		return false, fmt.Errorf("could not generate hash of filename '%s': %s", b.filename, err)
	}
	return hash != b.Hash(), nil
}

// Accessor functions for private variables
func (b Backend) Filename() string {
	return b.filename
}

func (b Backend) Hash() string {
	return b.hash
}
