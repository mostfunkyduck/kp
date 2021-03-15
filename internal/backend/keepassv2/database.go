package keepassv2

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
	g "github.com/tobischo/gokeepasslib/v3"
)

type Database struct {
	c.Database
	db *g.Database
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

func writeDB(db *g.Database, path string) error {
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

	path := d.SavePath()

	if err := writeDB(d.db, path); err != nil {
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

// open is a utility function to open the path stored as a database's SavePath
func (d *Database) open(opts t.Options) error {
	path := d.SavePath()

	dbReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open db file [%s]: %s", path, err)
	}

	err = g.NewDecoder(dbReader).Decode(d.db)
	if err != nil {
		// we need to swallow this error because it spews insane amounts of garbage for no good reason
		return fmt.Errorf("could not open database: is the password correct?")
	}
	if err := d.db.UnlockProtectedEntries(); err != nil {
		return fmt.Errorf("could not unlock protected entries: %s\n", err)
	}
	return nil
}

// Init will initialize the database.
func (d *Database) Init(opts t.Options) error {
	d.SetDriver(d)
	// the gokeepasslib always wants to start with a fresh DB
	// to use a DB on the filesystem, we will stomp this with a call to the appropriate decode function
	d.db = g.NewDatabase()
	// the v2 library prepopulates the db with a bunch of sample data, let's purge it
	if err := d.purge(); err != nil {
		return fmt.Errorf("failed to clean DB of sample data after initialization: %s", err)
	}
	d.SetSavePath(opts.DBPath)

	creds, err := d.credentials(opts)
	if err != nil {
		return fmt.Errorf("could not build credentials for given password: %s", err)
	}
	d.db.Credentials = creds

	// if the db already exists, open it, otherwise do an initial save and create the file
	if _, err := os.Stat(opts.DBPath); err == nil {
		if err := d.open(opts); err != nil {
			return err
		}
	} else {
		// this is a new db
		if err := d.Save(); err != nil {
			return fmt.Errorf("could not save newly created database: %s", err)
		}
	}

	d.SetCurrentLocation(d.Root())

	return nil
}

// purge will clean out all entries and groups from a database. Handle with care. Only provided to clean database of sample data on init
// should not be exposed to the user b/c that's far too dangerous
func (d *Database) purge() error {
	var failedRemovals string
	root := d.Root()
	for _, group := range root.Groups() {
		// removing the group from the root will orphan and effectively remove all the children of that group
		if err := root.RemoveSubgroup(group); err != nil {
			failedRemovals = failedRemovals + "," + group.Name()
		}
	}

	if failedRemovals != "" {
		return fmt.Errorf("failed to purge at least one group from the DB root, failed groups: [%s]", failedRemovals)
	}
	return nil
}

// credentials generates a the credentials for use in unlocking the database
func (d *Database) credentials(opts t.Options) (*g.DBCredentials, error) {
	// an empty password here is treated as valid, no special handling needed, it can be passed straight
	// to the keepassv2 library
	if opts.KeyPath == "" {
		// no key, we only need password creds
		creds := g.NewPasswordCredentials(opts.Password)
		return creds, nil
	}

	// There's a key, we need key/password creds
	keyReader, err := os.Open(opts.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not open key at path [%s]: %s", opts.KeyPath, err)
	}

	keyData, err := ioutil.ReadAll(keyReader)
	if err != nil {
		return nil, fmt.Errorf("could not read key that was opened from path [%s]: %s", opts.KeyPath, err)
	}

	creds, err := g.NewPasswordAndKeyDataCredentials(opts.Password, keyData)
	if err != nil {
		return creds, fmt.Errorf("could not build key/password credentials: %s", err)
	}

	return creds, nil
}

// Version returns the t.Version enum for this DB
func (d *Database) Version() t.Version {
	return t.V2
}
