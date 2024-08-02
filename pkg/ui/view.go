package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	switch m.currentPage {
	case home:
		return renderHome(m)
	case addConnection:
		return renderAddConnection(m)
	}
	return ""
}

func renderHome(m model) string {
	return appStyle.Render(m.list.View())
}

func renderAddConnection(m model) string {
	var b strings.Builder
	var button string
	// Render the text inputs
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	// TODO: is a button really necessary?
	if m.focusedInputIndex != len(m.inputs) {
		button = buttonStyle.Render("Add connection")
	} else {
		button = focusedButtonStyle.Render("Add connection")
	}

	popupContent := lipgloss.JoinVertical(lipgloss.Top, b.String(), button)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		popupStyle.Render(popupContent))

}
