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
	d.SetSavePath(options.DBPath)

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

	// technically, the savepath should not differ from options.DBPath,
	// but just incase something needs to be changed in the SavePath function, use that
	savePath := d.SavePath()
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

// Save will backup the DB, save it, then remove the backup is the save was successful
func (d *Database) Save() error {
	savePath := d.SavePath()

	if savePath == "" {
		return fmt.Errorf("no save path specified")
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
