package keepassv1

// wraps a v1 database with utility functions that allow it to be integrated
// into the shell.

import (
	"fmt"
	"io"
	"os"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Database struct {
	c.Database
	db *keepass.Database
}

// Init initializes the v1 database based on the provided options
func (d *Database) Init(options t.Options) error {
	var err error
	var keyReader io.Reader

	d.SetDriver(d)
	backend, err := c.InitBackend(options.DBPath)
	if err != nil {
		return fmt.Errorf("could not init backend: %s", err)
	}
	d.SetBackend(backend)
	if options.KeyPath != "" {
		keyReader, err = os.Open(options.KeyPath)
		if err != nil {
			return fmt.Errorf("could not open key file [%s]: %s\n", options.KeyPath, err)
		}
	}

	opts := &keepass.Options{
		Password:  options.Password,
		KeyFile:   keyReader,
		KeyRounds: options.KeyRounds,
	}

	savePath := d.Backend().Filename()
	if _, err := os.Stat(savePath); err == nil {
		dbReader, err := os.Open(savePath)
		if err != nil {
			return fmt.Errorf("could not open db file [%s]: %s\n", savePath, err)
		}

		db, err := keepass.Open(dbReader, opts)
		if err != nil {
			return fmt.Errorf("could not open database: %s\n", err)
		}
		d.db = db
	} else {
		db, err := keepass.New(opts)
		if err != nil {
			return fmt.Errorf("could not create new database with provided options: %s", err)
		}
		// need to set the internal db pointer before saving
		d.db = db

		if err := d.Save(); err != nil {
			return fmt.Errorf("could not save newly created database: %s", err)
		}
	}

	d.SetCurrentLocation(d.Root())
	return nil
}

// Root returns the DB root
func (d *Database) Root() t.Group {
	return WrapGroup(d.db.Root(), d)
}

// Save will backup the DB, save it, then remove the backup is the save was successful. it will also check to make sure the file has not changed.
func (d *Database) Save() error {
	savePath := d.Backend().Filename()

	if savePath == "" {
		return fmt.Errorf("no save path specified")
	}

	modified, err := d.Backend().IsModified()
	if err != nil {
		return fmt.Errorf("could not verify that the backend was unmodified: %s", err)
	}

	if modified {
		return fmt.Errorf("backend storage has been modified! please reopen before modifying to avoid corrupting or overwriting changes! (changes made since the last save will not be persisted)")
	}

	if err := d.Backup(); err != nil {
		return fmt.Errorf("could not back up database: %s", err)
	}

	w, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("could not open/create db save location [%s]: %s", savePath, err)
	}

	if err = d.db.Write(w); err != nil {
		return fmt.Errorf("error writing database to [%s]: %s", savePath, err)
	}

	if err := d.RemoveBackup(); err != nil {
		return fmt.Errorf("could not remove backup after saving: %s", err)
	}

	backend, err := c.InitBackend(savePath)
	if err != nil {
		return fmt.Errorf("error initializing new backend type after save: %s", err)
	}
	d.SetBackend(backend)
	return nil
}

func (d *Database) Raw() interface{} {
	return d.db
}

// Binary returns an OptionalWrapper with Present sent to false as v1 doesn't handle binaries
// through the database
func (d *Database) Binary(id int, name string) (t.OptionalWrapper, error) {
	return t.OptionalWrapper{
		Present: false,
	}, nil
}

// Version returns the t.Version enum representing this DB
func (d *Database) Version() t.Version {
	return t.V1
}
