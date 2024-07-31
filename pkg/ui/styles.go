package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle     = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)
