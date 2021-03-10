package keepassv2

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	g "github.com/tobischo/gokeepasslib/v3"
)

type Database struct {
	db              *g.Database
	currentLocation t.Group
	savePath        string
	options         t.Options
}

func NewDatabase(db *g.Database, savePath string, options t.Options) t.Database {
	dbWrapper := &Database{
		db:       db,
		savePath: savePath,
		options:  options,
	}
	dbWrapper.SetCurrentLocation(dbWrapper.Root())
	// the v2 library prepopulates the db with a bunch of sample data, let's purge it
	if _, err := os.Stat(savePath); err != nil {
		root := dbWrapper.Root()
		for _, group := range root.Groups() {
			if err := root.RemoveSubgroup(group); err != nil {
				fmt.Printf("could not purge initial subgroup %s: %s\n", group.Name(), err)
			}
		}
	}
	return dbWrapper
}

func (d *Database) Search(term *regexp.Regexp) (path []string, err error) {
	return d.Root().Search(term)
}

func (d *Database) Raw() interface{} {
	return d.db
}

func (d *Database) Root() t.Group {
	return &RootGroup{
		db:   d,
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
	if _, err := os.Stat(d.SavePath()); err != nil {
		// database path doesn't exist and doesn't need to be backed up
		return nil
	}

	data, err := ioutil.ReadFile(d.SavePath())
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(d.BackupPath(), data, 0644); err != nil {
		return err
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

	if err := db.LockProtectedEntries(); err != nil {
		panic(fmt.Sprintf("could not encrypt protected entries! database may be corrupted, save was not attempted: %s", err))
	}
	defer func() {
		if err := db.UnlockProtectedEntries(); err != nil {
			panic(fmt.Sprintf("could not decrypt protected entries! database may be corrupted, save was attempted: %s", err))
		}
	}()
	encoder := g.NewEncoder(f)
	if err := encoder.Encode(db); err != nil {
		return fmt.Errorf("could not write database: %s", err)
	}
	return nil
}
func (d *Database) Save() error {

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

func (d *Database) CurrentLocation() t.Group {
	return d.currentLocation
}

func (d *Database) SetCurrentLocation(g t.Group) {
	d.currentLocation = g
}

func (d *Database) SetSavePath(newPath string) {
	d.savePath = newPath
}

func (d *Database) SetOptions(opts t.Options) error {
	d.options = opts
	if opts.KeyReader == nil {
		d.db.Credentials = g.NewPasswordCredentials(opts.Password)
		return nil
	}

	// There's a key, if no password is specified, assume that the password is an empty string
	keyData, err := ioutil.ReadAll(opts.KeyReader)
	if err != nil {
		return err
	}
	creds, err := g.NewPasswordAndKeyDataCredentials(opts.Password, keyData)
	if err != nil {
		return err
	}

	d.db.Credentials = creds
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

// Binary returns a Value in an OptionalWrapper representing a binary
// because v2 stores half the metadata in the entry and half in the database,
// this function takes a 'Name' parameter so it can properly create the values
// Returns an empty Value (not even with a Name) if the binary doesn't exit,
// Returns a full Value if it does
func (d *Database) Binary(id int, name string) (t.OptionalWrapper, error) {
	binaryMeta := d.db.Content.Meta.Binaries
	meta := binaryMeta.Find(id)
	if meta == nil {
		return t.OptionalWrapper{
			Present: true,
			Value:   nil,
		}, nil
	}

	content, err := meta.GetContent()
	if err == io.EOF {
		content = ""
	} else if err != nil {
		return t.OptionalWrapper{Present: true}, err
	}
	return t.OptionalWrapper{
		Present: true,
		Value: c.NewValue(
			[]byte(content),
			name,
			false, false, false,
			t.BINARY,
		),
	}, nil
}
