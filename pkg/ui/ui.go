package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nezia1/ssh-manager/pkg/connection"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)
)

type keyMap struct {
	insertItem     key.Binding
	connect        key.Binding
	toggleHelpMenu key.Binding
	quit           key.Binding
}

type model struct {
	manager            connection.ConnectionManager
	list               list.Model
	keys               *keyMap
	selectedConnection *connection.Connection
}

func newKeyMap() *keyMap {
	return &keyMap{
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add connection"),
		),
		connect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "connect"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
		),
	}
}

func initialModel() model {
	var (
		cm   = connection.ConnectionManager{}
		list = list.New(cm.Items(), list.NewDefaultDelegate(), 0, 0)
		keys = newKeyMap()
	)

	list.Title = "Available connections"
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.insertItem,
			keys.connect,
			keys.toggleHelpMenu,
		}
	}

	return model{
		manager: cm,
		list:    list,
		keys:    keys,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.connect):
			selectedItem := m.list.SelectedItem().(connection.Item)
			m.selectedConnection = &selectedItem.Conn
			return m, tea.Quit
		case key.Matches(msg, m.keys.insertItem):
			password := "***REMOVED***"
			m.manager.AddConnection("infinity.usbx.me", "nezia", nil, &password)
			listCmd := m.list.SetItems(m.manager.Items())
			return m, tea.Batch(listCmd)
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func Start() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	m.(model).selectedConnection.StartSession()
}
