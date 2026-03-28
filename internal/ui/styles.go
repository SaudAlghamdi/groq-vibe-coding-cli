package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // purple
	secondaryColor = lipgloss.Color("#06B6D4") // cyan
	successColor   = lipgloss.Color("#10B981") // green
	warningColor   = lipgloss.Color("#F59E0B") // amber
	errorColor     = lipgloss.Color("#EF4444") // red
	mutedColor     = lipgloss.Color("#6B7280") // gray
	textColor      = lipgloss.Color("#E5E7EB") // light gray

	// Styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	UserPromptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	AssistantStyle = lipgloss.NewStyle().
			Foreground(textColor)

	ToolCallStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Italic(true)

	ToolResultStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(mutedColor).
			Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	DimStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)
