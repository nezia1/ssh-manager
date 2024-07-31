package ui

import "strings"

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

	// Render the text inputs
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	return appStyle.Render(b.String())
}
