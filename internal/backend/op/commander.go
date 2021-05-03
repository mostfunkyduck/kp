package op

import (
	"fmt"
	"os/exec"
	"strings"
)

// Wrapper around os/exec's command functions to allow dep injection
type Commander interface{
	// Command executes a command and returns the output/error
	Command(cmd string, stdin string) ([]byte, error)
	// SetSessionToken stores a session token for using in later calls to 1password
	SetSessionToken(string)
	// SessionToken retrieves the session token being used to call 1password
	SessionToken() string
}

type commander struct {
	token	string
}

func (c commander) Command(cmd string, stdin string) ([]byte, error) {
	cmd = cmd + " --session " + c.SessionToken() + " --cache"
	fmt.Println(cmd)
	cmdSlice := strings.Split(cmd, " ")
	cmdObj := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	if stdin != "" {
		cmdObj.Stdin = strings.NewReader(stdin)
	}
	return cmdObj.CombinedOutput()
}

func (c *commander) SetSessionToken(token string) {
	c.token = token
}

func (c commander) SessionToken() string {
	return c.token
}
