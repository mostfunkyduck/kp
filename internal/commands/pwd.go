package commands

import (
	"github.com/abiosoft/ishell"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func Pwd(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		db := shell.Get("db").(t.Database)
		path, err := db.Path()
		if err != nil {
			shell.Printf("could not retrieve current path: %s\n", err)
		}
		shell.Println(path)
	}
}
