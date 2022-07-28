package common

import (
	"fmt"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	"os"
	"regexp"
)

type Database struct {
	driver          t.Database
	currentLocation t.Group
	changed         bool
	backend         *Backend
}

// SetDriver sets pointer to the version of itself that can access child methods... FIXME this is a bit of a mind bender
func (d *Database) SetDriver(driver t.Database) {
	d.driver = driver
}

// SetBackend and Backend manage a cached hash representing the state of the backend
func (d *Database) SetBackend(backend *Backend) {
	d.backend = backend
}

func (d *Database) Backend() t.Backend {
	return d.backend
}

func (d *Database) lockPath() string {
	path := d.Backend().Filename()
	if path == "" {
		return path
	}

	return path + ".lock"
}

// Lock generates a lockfile for the given database
func (d *Database) Lock() error {
	path := d.lockPath()

	if path != "" {
		if _, err := os.Create(path); err != nil {
			return fmt.Errorf("could not create lock file at path [%s]: %s", path, err)
		}
	}
	return nil
}

// Unlock removes the lock file on the current savepath of the database
func (d *Database) Unlock() error {
	path := d.lockPath()

	if path != "" {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("could not remove lock file at path [%s]: %s", path, err)
		}
	}
	return nil
}

// Locked returns whether or not the lockfile exists
func (d *Database) Locked() bool {
	path := d.lockPath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (d *Database) Changed() bool {
	return d.changed
}

func (d *Database) SetChanged(changed bool) {
	d.changed = changed
}

var backupExtension = ".kpbackup"

// BackupPath returns the path to which a backup can be written or restored
func (d *Database) BackupPath() string {
	return d.Backend().Filename() + backupExtension
}

// Backup executes a backup, if the database exists, otherwise it will do nothing
func (d *Database) Backup() error {
	path := d.Backend().Filename()
	if _, err := os.Stat(path); err != nil {
		// database path doesn't exist and doesn't need to be backed up
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := os.WriteFile(d.BackupPath(), data, 0644); err != nil {
		return err
	}
	return nil
}

// RestoreBackup will restore a backup from the BackupPath() to the original file path. This will overwrite whatever's in the main location, handle with care
func (d *Database) RestoreBackup() error {
	backupPath := d.BackupPath()

	path := d.Backend().Filename()

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("no backup exists at [%s] for [%s], cannot restore", backupPath, path)
	}

	if err := os.Rename(backupPath, path); err != nil {
		return fmt.Errorf("could not rename '%s' to '%s': %s", backupPath, path, err)
	}

	return nil
}

// RemoveBackup will delete the backup database
func (d *Database) RemoveBackup() error {
	backupPath := d.BackupPath()

	if _, err := os.Stat(backupPath); err != nil {
		// no backup means we don't have to remove it either
		return nil
	}

	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("could not remove file '%s': %s", backupPath, err)
	}
	return nil
}

// CurrentLocation returns the group currently used as the user's shell location in the DB
func (d *Database) CurrentLocation() t.Group {
	return d.currentLocation
}

func (d *Database) SetCurrentLocation(g t.Group) {
	d.currentLocation = g
}

// Path will walk up the group hierarchy to determine the path to the current location
func (d *Database) Path() (string, error) {
	path, err := d.CurrentLocation().Path()
	if err != nil {
		return path, fmt.Errorf("could not find path to current location in database: %s", err)
	}
	return path, err
}

// Search looks through a database for an entry matching a given term
func (d *Database) Search(term *regexp.Regexp) (paths []string, err error) {
	return d.driver.Root().Search(term)
}

// SavePath is a shortcut for getting the backend's filename
func (d *Database) SavePath() string {
	return d.Backend().Filename()
}

// SetSavePath is a shortcut for setting the backend filename
func (d *Database) SetSavePath(path string) {
	backend := &Backend{
		filename: path,
		// hash is blank since this is considered a new file
		hash: "",
	}
	d.SetBackend(backend)
}
