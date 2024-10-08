package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func Start() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if m.(model).selectedConnection != nil {
		err := m.(model).selectedConnection.StartSession()
		if err != nil {
			log.Fatal(err)
		}
	}
}
