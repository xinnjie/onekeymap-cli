package views

import "github.com/charmbracelet/lipgloss"

// Color constants used across TUI components
const (
	blue   = lipgloss.Color("#5F87FF")
	white  = lipgloss.Color("#E4E4E4")
	red    = lipgloss.Color("#FF5F5F")
	yellow = lipgloss.Color("#FFFF5F")
	cyan   = lipgloss.Color("#5FFFFF")
	green  = lipgloss.Color("#5FFF5F")
	gray   = lipgloss.Color("#808080")

	IssuePaddingLeft = 2
	OptionPadding    = 2
)

// Style definitions used across TUI components
var (
	// Common styles
	//nolint:gochecknoglobals // style reused across TUI
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(blue).
			Padding(0, 1)

	//nolint:gochecknoglobals // style reused across TUI
	helpStyle = lipgloss.NewStyle().
			Foreground(gray).
			Padding(1, 1)

	// Validation report styles
	//nolint:gochecknoglobals // style reused across TUI
	summaryStyle = lipgloss.NewStyle().
			Foreground(white).
			Padding(0, 1)

	//nolint:gochecknoglobals // style reused across TUI
	errorHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(red).
				Padding(1, 1)

	//nolint:gochecknoglobals // style reused across TUI
	warningHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(yellow).
				Padding(1, 1)

	//nolint:gochecknoglobals // style reused across TUI
	issueStyle = lipgloss.NewStyle().
			PaddingLeft(IssuePaddingLeft)

	//nolint:gochecknoglobals // style reused across TUI
	keyStyle = lipgloss.NewStyle().
			Foreground(cyan)

	//nolint:gochecknoglobals // style reused across TUI
	actionStyle = lipgloss.NewStyle().
			Foreground(green)

	// Action detail styles
	//nolint:gochecknoglobals // style reused across TUI
	labelStyle = lipgloss.NewStyle().Bold(true)

	//nolint:gochecknoglobals // style reused across TUI
	supportedStyle = lipgloss.NewStyle().Foreground(green)

	//nolint:gochecknoglobals // style reused across TUI
	notSupportedStyle = lipgloss.NewStyle().Foreground(red)
)
