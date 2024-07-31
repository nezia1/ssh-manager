package ui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nezia1/ssh-manager/pkg/connection"
)

var (
	appStyle     = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

type keyMap struct {
	insertItem     key.Binding
	connect        key.Binding
	toggleHelpMenu key.Binding
	quit           key.Binding
}

type page int

type model struct {
	manager            connection.ConnectionManager
	list               list.Model
	keys               *keyMap
	selectedConnection *connection.Connection
	inputs             []textinput.Model
	focusedInputIndex  int
	currentPage        page
}

const (
	home page = iota
	addConnection
)

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
		cm     = connection.ConnectionManager{}
		list   = list.New(cm.Items(), list.NewDefaultDelegate(), 0, 0)
		keys   = newKeyMap()
		inputs = make([]textinput.Model, 2)
	)

	for i := range inputs {
		t := textinput.New()
		switch i {
		case 0:
			t.Placeholder = "SSH string (format user@host:port)"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.Focus()
		case 1:
			t.Placeholder = "Password (optional)"
			t.PromptStyle = blurredStyle
			t.TextStyle = blurredStyle
		}
		inputs[i] = t
	}

	list.Title = "Available connections"
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.insertItem,
			keys.connect,
			keys.toggleHelpMenu,
		}
	}

	return model{
		manager:           cm,
		list:              list,
		keys:              keys,
		currentPage:       home,
		inputs:            inputs,
		focusedInputIndex: 0,
	}
}

func renderHome(m model) string {
	return appStyle.Render(m.list.View())
}

func renderAddConnection(m model) string {
	var b strings.Builder

	// Render the text inputs
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	return appStyle.Render(b.String())
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentPage {
	case home:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.connect):
				selectedItem := m.list.SelectedItem().(connection.Item)
				m.selectedConnection = &selectedItem.Conn
				return m, tea.Quit
			case key.Matches(msg, m.keys.insertItem):
				m.currentPage = addConnection
			case key.Matches(msg, m.keys.quit):
				return m, tea.Quit
			}
		case tea.WindowSizeMsg:
			h, v := appStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		}
	case addConnection:
		switch msg := msg.(type) {
		case tea.KeyMsg:

			switch msg.String() {
			case "tab", "shift+tab", "up", "down", "enter":
				s := msg.String()

				if s == "enter" {
					return m, tea.Quit
				}

				if s == "tab" || s == "down" {
					m.focusedInputIndex = (m.focusedInputIndex + 1) % len(m.inputs)
				} else {
					m.focusedInputIndex = (m.focusedInputIndex - 1 + len(m.inputs)) % len(m.inputs)
				}

				cmds := make([]tea.Cmd, len(m.inputs))

				for i := range m.inputs {
					if i == m.focusedInputIndex {
						cmds[i] = m.inputs[i].Focus()
						m.inputs[i].PromptStyle = focusedStyle
						m.inputs[i].TextStyle = focusedStyle
						continue
					}

					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = blurredStyle
					m.inputs[i].TextStyle = blurredStyle
				}

				return m, tea.Batch(cmds...)
			}
		}
	}

	var listCmd, inputsCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	inputsCmd = m.updateInputs(msg)

	return m, tea.Batch(listCmd, inputsCmd)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.currentPage {
	case home:
		return renderHome(m)
	case addConnection:
		return renderAddConnection(m)
	}
	return ""
}

func Start() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if m.(model).selectedConnection != nil {
		m.(model).selectedConnection.StartSession()
	}
}
