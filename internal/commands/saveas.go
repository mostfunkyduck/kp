package commands

import (
	"github.com/abiosoft/ishell"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

func SaveAs(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 1)
		if !ok {
			shell.Println(errString)
			return
		}

		savePath := c.Args[0]

		if !confirmOverwrite(shell, savePath) {
			shell.Println("not overwriting existing file")
			return
		}

		db := shell.Get("db").(t.Database)

		oldPath := db.SavePath()

		db.SetSavePath(savePath)
		if err := db.Save(); err != nil {
			shell.Printf("could not save database: %s\n", err)
		}

		db.SetSavePath(oldPath)
	}
}
