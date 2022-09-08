package task

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
)

var (
	// TextStyle is the default text style
	TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dddddd")).Render

	// ErrorStyle is the default error text
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd0000")).Bold(true).Render

	Spinner = spinner.Spinner{
		Frames: []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"},
		FPS: time.Second/12,
	}
)
