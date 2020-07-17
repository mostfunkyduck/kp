package keepassv1

// wraps a v1 database with utility functions that allow it to be integrated
// into the shell.

import (
	"fmt"
	"os"

	k "github.com/mostfunkyduck/kp/keepass"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

type Database struct {
	currentLocation k.Group
	db              *keepass.Database
	savePath        string
}

var backupExtension = ".kpbackup"

func NewDatabase(db *keepass.Database, savePath string) k.Database {
	rv := &Database{
		currentLocation: NewGroup(db.Root()),
		db:              db,
		savePath:        savePath,
	}
	return rv
}

// Root returns the DB root
func (d *Database) Root() k.Group {
	return NewGroup(d.db.Root())
}

// Backup will create a backup of the current database to a temporary location
// in case saving the database causes some kind of corruption
func (d *Database) Backup() error {
	backupPath := d.SavePath() + backupExtension
	w, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("could not open file '%s': %s", backupPath, err)
	}

	if err := d.db.Write(w); err != nil {
		return fmt.Errorf("could not write to file '%s': %s", backupPath, err)
	}
	return nil
}

// RemoveBackup removes a temporary backup file
func (d *Database) RemoveBackup() error {
	backupPath := d.SavePath() + backupExtension
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("could not remove backup file '%s': %s", backupPath, err)
	}
	return nil
}

// SavePath returns the current save location for the DB
func (d *Database) SavePath() string {
	return d.savePath
}

func (d *Database) SetSavePath(newPath string) {
	d.savePath = newPath
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

func (d *Database) SetOptions(o k.Options) error {
	opts := &keepass.Options{
		Password: o.Password,
		KeyFile:  o.KeyReader,
	}
	if err := d.db.SetOpts(opts); err != nil {
		return fmt.Errorf("could not set DB options: %s", err)
	}
	return nil
}

func (d *Database) CurrentLocation() k.Group {
	return d.currentLocation
}

func (d *Database) SetCurrentLocation(g k.Group) {
	d.currentLocation = g
}

func (d *Database) Raw() interface{} {
	return d.db
}

// Path will walk up the group hierarchy to determine the path to the current location
func (d *Database) Path() (fullPath string) {
	group := d.CurrentLocation()
	return group.Path()
}
