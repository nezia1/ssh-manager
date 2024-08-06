package ui

import (
	"log"
	"strconv"
	"strings"

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

	// initialize text inputs
	sshInput := textinput.New()
	sshInput.Placeholder = "SSH string (format user@host:port)"
	sshInput.PromptStyle = focusedStyle
	sshInput.TextStyle = focusedStyle
	sshInput.Focus()

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password (optional)"
	passwordInput.PromptStyle = blurredStyle
	passwordInput.TextStyle = blurredStyle
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'

	inputs[0] = sshInput
	inputs[1] = passwordInput

	// initialize list
	list.Title = "Available connections"
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.insertItem,
			keys.connect,
			keys.toggleHelpMenu,
		}
	}

	// initialize model
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
	cmds := []tea.Cmd{}

	switch m.currentPage {
	case home:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.connect):
				selectedItem := m.list.SelectedItem().(connection.Item)
				m.selectedConnection = &selectedItem.Conn
				cmds = append(cmds, tea.Quit)
			case key.Matches(msg, m.keys.insertItem):
				m.currentPage = addConnection
			case key.Matches(msg, m.keys.quit):
				cmds = append(cmds, tea.Quit)
			}
		}
	case addConnection:
		cmds = append(cmds, m.updateAddConnection(msg)...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResizing(msg)
	case connection.ConnectionsFetchedMsg:
		m.manager = msg.FetchedManager
		cmds = append(cmds, m.list.SetItems(m.manager.Items()))
	}

	// update the list and inputs with the current message
	var listCmd, inputsCmd tea.Cmd

	m.list, listCmd = m.list.Update(msg)
	inputsCmd = m.updateInputs(msg)

	cmds = append(cmds, listCmd, inputsCmd)

	return m, tea.Batch(cmds...)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.manager.FetchConnections, textinput.Blink)
}

// updateInputs updates the text inputs when typing into them.
func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// updateAddConnection handles the key presses when on the add connection page. This includes navigation, tabbing between inputs, submitting the form, and adding the connection to the manager.
//
// It returns a slice of commands to be executed.
func (m *model) updateAddConnection(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down", "enter":
			cmds = append(cmds, m.handleInputNavigation(msg.String())...)
			// TODO: add I/O with config file
			if msg.String() == "enter" {
				host, username, password, port, err := m.parseConnectionInput()
				if err != nil {
					log.Fatal(err)
				}
				m.manager.AddConnection(host, username, port, &password)
				m.currentPage = home

				cmd := m.list.SetItems(m.manager.Items())

				cmds = append(cmds, cmd)
			}

		}
		return nil
	}
	return cmds
}

// parseConnectionInput parses the input from the text inputs (ssh string and password).
//
// Returns all the necessary data to create a SSH connection, or an error if the input is invalid in any way.
func (m *model) parseConnectionInput() (host, username, password string, port int, err error) {
	parts := strings.Split(m.inputs[0].Value(), "@")

	username = parts[0]
	parts = strings.Split(parts[1], ":")

	host = parts[0]

	if len(parts) == 1 {
		port = 0
	} else {
		port, err = strconv.Atoi(parts[1])
	}

	if err != nil {
		// TODO: show error
		return "", "", "", 0, err
	}

	if m.inputs[1].Value() != "" {
		password = m.inputs[1].Value()
	}

	return host, username, password, port, nil
}

// handleInputNavigation handles the navigation between the text inputs and the button.
//
// It returns a slice of commands to be executed.
func (m *model) handleInputNavigation(key string) []tea.Cmd {
	var cmds []tea.Cmd
	// adding one to account for the button
	if key == "tab" || key == "down" {
		m.focusedInputIndex = (m.focusedInputIndex + 1) % (len(m.inputs) + 1)
	} else {
		m.focusedInputIndex = (m.focusedInputIndex - 1 + (len(m.inputs)) + 1) % (len(m.inputs) + 1)
	}

	for i := range m.inputs {
		if i == m.focusedInputIndex {
			cmds = append(cmds, m.inputs[i].Focus())
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}

		// we need to check if we're not on the button
		if i < len(m.inputs) {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = blurredStyle
			m.inputs[i].TextStyle = blurredStyle
		}
	}
	return cmds
}

// handleResizing handles the resizing of the window. It updates the width and height of the components globally.
func (m *model) handleResizing(msg tea.WindowSizeMsg) {
	h, v := appStyle.GetFrameSize()
	m.width = msg.Width - h
	m.height = msg.Height - v
	m.list.SetSize(m.width, m.height)
	popupStyle = popupStyle.Width(m.width / 2).Height(m.height / 2)

	for i := range m.inputs {
		m.inputs[i].Width = m.width / 4
	}
}
