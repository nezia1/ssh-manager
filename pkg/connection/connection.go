package connection

import (
	"fmt"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"github.com/charmbracelet/bubbles/list"
)

type Connection struct {
	Username string
	Host     string
	Port     *int
	Password *string // stored as a pointer to allow for nil values (TODO add encryption)
}

func (c Connection) SSHCommand() string {
	var command []string
	if c.Password == nil {
		command = []string{"ssh", fmt.Sprintf("%s@%s", c.Username, c.Host)}
	} else {
		command = []string{"sshpass", "-p", *c.Password, "ssh", fmt.Sprintf("%s@%s", c.Username, c.Host)}
	}

	if c.Port != nil {
		command = append(command, "-p", fmt.Sprintf("%d", *c.Port))
	}

	return shellescape.QuoteCommand(command)
}

type Item struct {
	Conn Connection
}

func (i Item) Title() string {
	var title strings.Builder

	fmt.Fprintf(&title, "%s@%s", i.Conn.Username, i.Conn.Host)
	if i.Conn.Port != nil {
		fmt.Fprintf(&title, ":%d", i.Conn.Port)
	}
	return title.String()

}

func (i Item) Description() string {
	return i.Conn.SSHCommand()
}

func (i Item) FilterValue() string {
	return i.Conn.Host
}

type ConnectionManager struct {
	Connections []Connection
}

func (cm *ConnectionManager) AddConnection(host string, user string, port *int, password *string) {
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
