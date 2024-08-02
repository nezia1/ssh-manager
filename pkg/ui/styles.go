package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle     = lipgloss.NewStyle()
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	popupStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Align(lipgloss.Center, lipgloss.Center)
)
