package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nezia1/ssh-manager/pkg/connection"
)

type keyMap struct {
	insertItem     key.Binding
	connect        key.Binding
	toggleHelpMenu key.Binding
	quit           key.Binding
}

type page int

const (
	home page = iota
	addConnection
)

type model struct {
	manager            connection.ConnectionManager
	list               list.Model
	keys               *keyMap
	selectedConnection *connection.Connection
	inputs             []textinput.Model
	focusedInputIndex  int
	currentPage        page
	width              int
	height             int
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
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		m.list.SetSize(m.width, m.height)
		popupStyle = popupStyle.Width(m.width / 2).Height(m.height / 2)
	}

	var listCmd, inputsCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	inputsCmd = m.updateInputs(msg)

	for i := range m.inputs {
		m.inputs[i].Width = m.width / 4

	}
	return m, tea.Batch(listCmd, inputsCmd)
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
