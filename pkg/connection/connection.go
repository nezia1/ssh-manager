package connection

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/x/term"
	"golang.org/x/crypto/ssh"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

func (c Connection) SSHCommand() (string, []string, error) {
	var command string
	var args []string

	if !c.IsPassword {
		command = "ssh"
	} else {
		command = "sshpass"
		password, err := c.Password()
		if err != nil {
			return "", nil, fmt.Errorf("unable to get password for ssh connection %v: %v", c.Host, err)
		}
		args = append(args, "-p", password, "ssh")
	}

	args = append(args, fmt.Sprintf("%s@%s", c.Username, c.Host))
	args = append(args, "-p", fmt.Sprintf("%d", c.Port))

	return command, args, nil
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

func (i Item) Description() string {
	var description string
	if i.Conn.IsPassword {
		description = "Password connection"
	} else {
		description = "SSH key connection"
	}

	return description
}

func (i Item) FilterValue() string {
	return i.Conn.Host
}

type ConnectionManager struct {
	Connections []Connection
}

func (cm *ConnectionManager) AddConnection(host string, user string, port int, password *string) error {
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
			return fmt.Errorf("failed to store password after adding new connection: %v", err)
		}

		connection.IsPassword = true
	}

	cm.Connections = append(cm.Connections, connection)

	err := cm.SaveToDisk()

	if err != nil {
		return fmt.Errorf("failed to save to disk after adding new connection: %v", err)
	}

	return nil
}

func (cm *ConnectionManager) DeleteConnection(index int) error {
	// delete password from pass if applicable
	connectionToDelete := cm.Connections[index]
	if connectionToDelete.IsPassword {
		err := connectionToDelete.RemovePassword()
		if err != nil {
			return fmt.Errorf("failed to remove password when deleting connection: %v", err)
		}
	}

	// delete connection
	cm.Connections = append(cm.Connections[:index], cm.Connections[index+1:]...)

	err := cm.SaveToDisk()

	if err != nil {
		return fmt.Errorf("failed to save to disk: %v", err)
	}
	return nil
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
		return fmt.Errorf("unable to get user home directory: %v", err)
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
		return fmt.Errorf("unable to start ssh connection: %v", err)
	}

	session, err := client.NewSession()

	if err != nil {
		return fmt.Errorf("unable to start ssh session: %v", err)
	}

	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing (needed in raw mode, otherwise typed characters are invisible since sent directly to the ssh session)
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fd := os.Stdin.Fd()

	// this is needed so that special characters (escape sequences such as CTRL-C) can be sent directly to the session
	oldState, err := term.MakeRaw(fd)

	if err != nil {
		return fmt.Errorf("failed to set terminal into raw mode: %v", err)
	}

	defer term.Restore(fd, oldState)

	// handles resize asynchronously
	go handleResize(session)

	w, h, err := term.GetSize(fd)

	if err != nil {
		return fmt.Errorf("cannot get terminal size: %v", err)
	}

	// request a pseudo-terminal
	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %v", err)
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

// handleResize creates a channel and listens to it for SIGWINCH. It handles resizing the ssh session, as we need to explicitely inform it when our terminal window size changes.
//
// Meant to be used as a goroutine.
func handleResize(session *ssh.Session) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGWINCH)
	for {
		<-signals

		fd := os.Stdin.Fd()
		cols, rows, err := term.GetSize(fd)

		if err != nil {
			fmt.Printf("failed to get terminal size: %v", err)
		}

		if err := session.WindowChange(rows, cols); err != nil {
			fmt.Printf("error resizing terminal: %v", err)
		}
	}
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
	password, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		return "", err
	}

	return string(password), nil
}
