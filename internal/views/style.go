package views

import "github.com/charmbracelet/lipgloss"

// Color constants used across TUI components
const (
	blue      = lipgloss.Color("12")
	white     = lipgloss.Color("7")
	red       = lipgloss.Color("9")
	yellow    = lipgloss.Color("11")
	cyan      = lipgloss.Color("14")
	green     = lipgloss.Color("10")
	gray      = lipgloss.Color("8")
	purple    = lipgloss.Color("#7C3AED")
	lightGray = lipgloss.Color("#666666")

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

	// Telemetry prompt styles
	//nolint:gochecknoglobals // style reused across TUI
	telemetryTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(purple).
				Padding(1, 0)

	//nolint:gochecknoglobals // style reused across TUI
	questionStyle = lipgloss.NewStyle().
			Padding(1, 0)

	//nolint:gochecknoglobals // style reused across TUI
	optionStyle = lipgloss.NewStyle().
			Padding(0, OptionPadding)

	//nolint:gochecknoglobals // style reused across TUI
	selectedOptionStyle = lipgloss.NewStyle().
				Padding(0, OptionPadding).
				Background(purple).
				Foreground(lipgloss.Color("#FFFFFF"))

	//nolint:gochecknoglobals // style reused across TUI
	telemetryHelpStyle = lipgloss.NewStyle().
				Foreground(lightGray).
				Padding(1, 0)

	// Action detail styles
	//nolint:gochecknoglobals // style reused across TUI
	labelStyle = lipgloss.NewStyle().Bold(true)

	//nolint:gochecknoglobals // style reused across TUI
	supportedStyle = lipgloss.NewStyle().Foreground(green)

	//nolint:gochecknoglobals // style reused across TUI
	notSupportedStyle = lipgloss.NewStyle().Foreground(red)
)
