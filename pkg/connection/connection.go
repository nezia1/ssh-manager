package connection

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultPort = 22
)

type Connection struct {
	Username string
	Host     string
	Port     int
	Password *string // stored as a pointer to allow for nil values (TODO add encryption)
}

func (c Connection) SSHCommand() (string, []string) {
	var command string
	var args []string

	if c.Password == nil {
		command = "ssh"
	} else {
		command = "sshpass"
		args = append(args, "-p", *c.Password, "ssh")
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
		Password: password,
	}

	cm.Connections = append(cm.Connections, connection)
}

func (cm ConnectionManager) Items() []list.Item {
	items := []list.Item{}
	for _, conn := range cm.Connections {
		items = append(items, Item{Conn: conn})
	}
	return items
}

func (c Connection) StartSession() {
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(*c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

	// Request a pseudo-terminal
	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		log.Fatal("request for pseudo terminal failed: ", err)
	}

	session.Stdout = os.Stdout
	session.Stdin = os.Stdin
	session.Stderr = os.Stderr

	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatal(err)
	}

	// Wait for remote shell to close
	err = session.Wait()
	if err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
}
