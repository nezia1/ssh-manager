package connection

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	DefaultPort = 22
)

type Connection struct {
	Username   string
	Host       string
	Port       int
	IsPassword bool
}

func (c Connection) SSHCommand() (string, []string) {
	var command string
	var args []string

	if !c.IsPassword {
		command = "ssh"
	} else {
		command = "sshpass"
		password, err := c.Password()
		if err != nil {
			log.Fatal(err)
		}
		args = append(args, "-p", password, "ssh")
	}

	args = append(args, fmt.Sprintf("%s@%s", c.Username, c.Host))
	args = append(args, "-p", fmt.Sprintf("%d", c.Port))

	return command, args
}

type Item struct {
	Conn Connection
}

func (i Item) Title() string {
	var title strings.Builder

	fmt.Fprintf(&title, "%s@%s", i.Conn.Username, i.Conn.Host)
	if i.Conn.Port == 0 {
		fmt.Fprintf(&title, ":%d", i.Conn.Port)
	}
	return title.String()

}

// TODO: avoid showing password in description
func (i Item) Description() string {
	command, args := i.Conn.SSHCommand()
	return fmt.Sprintf("%s %s", command, strings.Join(args, " "))
}

func (i Item) FilterValue() string {
	return i.Conn.Host
}

type ConnectionManager struct {
	Connections []Connection
}

func (cm *ConnectionManager) AddConnection(host string, user string, port int, password *string) {
	if port == 0 {
		port = DefaultPort
	}

	connection := Connection{
		Username: user,
		Host:     host,
		Port:     port,
	}

	if password != nil {
		err := connection.StorePassword(*password)

		if err != nil {
			log.Fatal(err)
		}

		connection.IsPassword = true
	}

	cm.Connections = append(cm.Connections, connection)

	err := cm.SaveToDisk()

	if err != nil {
		log.Fatal(fmt.Printf("failed to save to disk: %v", err))
	}
}

func (cm ConnectionManager) Items() []list.Item {
	items := []list.Item{}
	for _, conn := range cm.Connections {
		items = append(items, Item{Conn: conn})
	}
	return items
}

// StartSession starts a new SSH session with the given connection. It will prompt for a password if one was provided, otherwise, it will try to find all available keys in the default paths.
// TODO: allow custom key paths
func (c Connection) StartSession() error {
	var authMethods []ssh.AuthMethod

	if c.IsPassword {
		password, err := c.Password()

		if err != nil {
			return err
		}

		authMethods = append(authMethods, ssh.Password(password))
	}

	// try to find all available keys in the default paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	keyPaths := []string{
		fmt.Sprintf("%s/.ssh/id_rsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_dsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_ecdsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
	}

	for _, keyPath := range keyPaths {
		newAuthMethod, err := publicKeyFile(keyPath)
		if err != nil {
			continue
		}

		authMethods = append(authMethods, newAuthMethod)
	}

	config := &ssh.ClientConfig{
		User:            c.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: add host key verification
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), config)

	if err != nil {
		log.Fatal(err)
	}

	session, err := client.NewSession()

	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// TODO: fix escape sequences such as CTRL-L and arrow keys echoing instead of being interpreted
	// request a pseudo-terminal
	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		log.Fatal("request for pseudo terminal failed: ", err)
	}

	session.Stdout = os.Stdout
	session.Stdin = os.Stdin
	session.Stderr = os.Stderr

	// start remote shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %v", err)
	}

	// wait for remote shell to close
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("remote shell exited with error: %v", err)
	}

	return nil
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err == nil {
		return ssh.PublicKeys(signer), nil
	}

	fmt.Printf("Enter passphrase for key %s:", file)
	passphrase, err := readPassphrase()

	if err != nil {
		return nil, err
	}

	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))

	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func readPassphrase() (string, error) {
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(password), nil
}
