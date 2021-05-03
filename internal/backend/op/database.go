package op

// wraps a v1 database with utility functions that allow it to be integrated
// into the shell.

import (
	"fmt"

	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

type Database struct {
	c.Database
	OPDriver	*Driver
}

func (d *Database) Init(options t.Options) error {
	driver, err := NewDriver(Credentials {
		Username: "fake", // we'll get there, we'll get there
		Password: options.Password,
		Account: "whitebox", // no, seriously, we'll get there
	}, &commander{})
	if err != nil {
		return fmt.Errorf("got error trying to start driver: %s", err)
	}
	err = driver.SignIn()
	if err != nil {
		return fmt.Errorf("couldn't sign in: %s", err)
	}
	d.OPDriver = &driver
	d.SetDriver(d)
	d.SetCurrentLocation(d.Root())
	return nil
}

func (d *Database) Root() t.Group {
	return NewGroup("", d, d.OPDriver)
}

func (d *Database) Save() error {
	return fmt.Errorf("not implemented")
}

func (d *Database) Raw() interface{} {
	return d
}

func (d *Database) Binary(id int, name string) (t.OptionalWrapper, error) {
	return t.OptionalWrapper{}, fmt.Errorf("not implemented")
}

func (d *Database) Version() t.Version {
	return t.V1
}

func (d *Database) Locked() bool {
	return false
}
