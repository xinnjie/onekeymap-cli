package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TelemetryPromptModel represents the telemetry consent prompt
type TelemetryPromptModel struct {
	choice   int  // 0 = allow, 1 = deny
	selected bool // whether user has made a selection
	quitting bool
}

// NewTelemetryPrompt creates a new telemetry prompt model
func NewTelemetryPrompt() TelemetryPromptModel {
	return TelemetryPromptModel{
		choice:   0, // Default to allow telemetry
		selected: false,
		quitting: false,
	}
}

func (m TelemetryPromptModel) Init() tea.Cmd {
	return nil
}

func (m TelemetryPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.choice > 0 {
				m.choice--
			}
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.choice < 1 {
				m.choice++
			}
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))):
			m.selected = true
			return m, tea.Quit
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m TelemetryPromptModel) View() string {
	if m.quitting {
		return ""
	}

	title := telemetryTitleStyle.Render("🔒 Privacy Notice")

	question := questionStyle.Render(
		"Onekeymap CLI would like to collect anonymous usage data to help improve the tool.\n" +
			"This includes information about:\n" +
			"  • Unknown editor commands, so we can provide prioritized support\n" +
			"  • Commands usage\n" +
			"\n" +
			"No personal data, file contents, or keystrokes are collected.\n" +
			"More info: https://github.com/onekeymap/onekeymap-cli/blob/main/docs/telemetry.md\n" +
			"Would you like to help?",
	)

	var options string
	if m.choice == 0 {
		options += selectedOptionStyle.Render("▶ Yes, enable telemetry") + "\n"
		options += optionStyle.Render("  No, keep it disabled")
	} else {
		options += optionStyle.Render("  Yes, enable telemetry") + "\n"
		options += selectedOptionStyle.Render("▶ No, keep it disabled")
	}

	help := telemetryHelpStyle.Render("Use ↑/↓ or j/k to select, Enter to confirm, q to quit")

	return fmt.Sprintf("%s\n%s\n%s\n\n%s", title, question, options, help)
}

// GetChoice returns the user's choice: true for enabled, false for disabled
func (m TelemetryPromptModel) GetChoice() bool {
	return m.choice == 0
}

// WasSelected returns whether the user made a selection (vs quitting)
func (m TelemetryPromptModel) WasSelected() bool {
	return m.selected
}
