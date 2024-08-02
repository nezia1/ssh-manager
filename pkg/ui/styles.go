package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle     = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	popupStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Align(lipgloss.Center, lipgloss.Center)
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("12")).
			Padding(0, 3).
			MarginTop(1)
	focusedButtonStyle = buttonStyle.Background(lipgloss.Color("5"))
)
