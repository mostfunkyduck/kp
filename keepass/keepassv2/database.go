package keepassv2

import (
	"fmt"
	k "github.com/mostfunkyduck/kp/keepass"
	g "github.com/tobischo/gokeepasslib/v3"
	"os"
)

type Database struct {
	db              *g.Database
	currentLocation k.Group
	savePath        string
	options         k.Options
}

func NewDatabase(db *g.Database, savePath string, options k.Options) k.Database {
	dbWrapper := &Database{
		db:       db,
		savePath: savePath,
		options:  options,
	}
	dbWrapper.SetCurrentLocation(dbWrapper.Root())
	return dbWrapper
}

//KeepassWrapper
func (d *Database) Raw() interface{} {
	return d.db
}

func (d *Database) Root() k.Group {
	return &RootGroup{
		root: d.db.Content.Root,
	}
}

var backupExtension = ".kpbackup"

func (d *Database) SavePath() string {
	return d.savePath
}
func (d *Database) BackupPath() string {
	return d.SavePath() + backupExtension
}

func (d *Database) Backup() error {
	backupPath := d.BackupPath()
	if err := writeDb(d.db, backupPath); err != nil {
		return fmt.Errorf("could not back up database: %s", err)
	}
	return nil
}

func (d *Database) RestoreBackup() error {
	if err := os.Rename(d.BackupPath(), d.SavePath()); err != nil {
		return fmt.Errorf("could not rename '%s' to '%s': %s", d.BackupPath(), d.SavePath(), err)
	}
	return nil
}

func (d *Database) RemoveBackup() error {
	if err := os.Remove(d.BackupPath()); err != nil {
		return fmt.Errorf("could not remove file '%s': %s", d.BackupPath(), err)
	}
	return nil
}

func writeDb(db *g.Database, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not open file '%s': %s", path, err)
	}

	encoder := g.NewEncoder(f)
	if err := encoder.Encode(db); err != nil {
		return fmt.Errorf("could not write database: %s", err)
	}
	return nil
}
func (d *Database) Save() error {
	if err := d.db.LockProtectedEntries(); err != nil {
		return fmt.Errorf("could not lock (encrypt) protected entries: %s", err)
	}
	defer func() {
		if err := d.db.UnlockProtectedEntries(); err != nil {
			fmt.Printf("error unlocking protected entries, database may be corrupted")
		}
	}()

	if err := d.Backup(); err != nil {
		return fmt.Errorf("could not back up database: %s", err)
	}

	if err := writeDb(d.db, d.SavePath()); err != nil {
		// TODO put this call in v1 too
		if backupErr := d.RestoreBackup(); backupErr != nil {
			return fmt.Errorf("could not save database: %s. also could not restore backup after failed save: %s", err, backupErr)
		}
		return fmt.Errorf("could not save database: %s", err)
	}
	if err := d.RemoveBackup(); err != nil {
		return fmt.Errorf("could not remove backup after successful save: %s", err)
	}
	return nil
}

func (d *Database) CurrentLocation() k.Group {
	return d.currentLocation
}

func (d *Database) SetCurrentLocation(g k.Group) {
	d.currentLocation = g
}

func (d *Database) SetSavePath(newPath string) {
	d.savePath = newPath
}

func (d *Database) SetOptions(opts k.Options) error {
	d.options = opts
	return nil
}

// Path will walk up the group hierarchy to determine the path to the current location
func (d *Database) Path() (string, error) {
	path, err := d.CurrentLocation().Path()
	if err != nil {
		return path, fmt.Errorf("could not find path to current location in database: %s", err)
	}
	return path, err
}
