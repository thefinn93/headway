package task

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// TextStyle is the default text style
	TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dddddd")).Render

	// ErrorStyle is the default error text
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd0000")).Bold(true).Render

	padding = "  "
)
