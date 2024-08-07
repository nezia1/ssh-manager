package connection

import (
	"fmt"
	"os/exec"
	"strings"
)

// Store the password in pass
func (c *Connection) StorePassword(password string) error {
	cmd := exec.Command("pass", "insert", "-e", fmt.Sprintf("%s@%s", c.Username, c.Host))
	cmd.Stdin = strings.NewReader(password)

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to store password: %v", err)
	}

	return nil
}

func (c *Connection) Password() (string, error) {
	cmd := exec.Command("pass", fmt.Sprintf("%s@%s", c.Username, c.Host))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}

	return string(out), nil
}
