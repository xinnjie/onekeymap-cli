package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var (
	blue          = lipgloss.Color("12")
	white         = lipgloss.Color("7")
	red           = lipgloss.Color("9")
	yellow        = lipgloss.Color("11")
	cyan          = lipgloss.Color("14")
	green         = lipgloss.Color("10")
	gray          = lipgloss.Color("8")
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(blue).Padding(0, 1)
	summaryStyle  = lipgloss.NewStyle().Foreground(white).Padding(0, 1)
	errorHeader   = lipgloss.NewStyle().Bold(true).Foreground(red).Padding(1, 1)
	warningHeader = lipgloss.NewStyle().Bold(true).Foreground(yellow).Padding(1, 1)
	issueStyle    = lipgloss.NewStyle().PaddingLeft(2)
	keyStyle      = lipgloss.NewStyle().Foreground(cyan)
	actionStyle   = lipgloss.NewStyle().Foreground(green)
	helpStyle     = lipgloss.NewStyle().Foreground(gray).Padding(1, 1)
)

type ValidationReportModel struct {
	report   *keymapv1.ValidationReport
	viewport viewport.Model
	ready    bool
}

func NewValidationReportModel(report *keymapv1.ValidationReport) ValidationReportModel {
	return ValidationReportModel{
		report: report,
	}
}

func (m ValidationReportModel) Init() tea.Cmd {
	return nil
}

func (m ValidationReportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(m.renderReport())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ValidationReportModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m ValidationReportModel) headerView() string {
	return titleStyle.Render("Validation Report")
}

func (m ValidationReportModel) footerView() string {
	return helpStyle.Render("Press Enter or q to continue.")
}

func (m ValidationReportModel) renderReport() string {
	var b strings.Builder

	// Summary
	s := m.report.Summary
	b.WriteString(summaryStyle.Render(
		fmt.Sprintf("Source: %s | Mappings Processed: %d | Succeeded: %d",
			m.report.SourceEditor, s.MappingsProcessed, s.MappingsSucceeded),
	))
	b.WriteString("\n")

	// Errors
	if len(m.report.Issues) > 0 {
		b.WriteString(errorHeader.Render(fmt.Sprintf("Errors (%d)", len(m.report.Issues))))
		for _, issue := range m.report.Issues {
			b.WriteString(renderIssue(issue))
		}
	}

	// Warnings
	if len(m.report.Warnings) > 0 {
		b.WriteString(warningHeader.Render(fmt.Sprintf("Warnings (%d)", len(m.report.Warnings))))
		for _, warning := range m.report.Warnings {
			b.WriteString(renderIssue(warning))
		}
	}

	return b.String()
}

func renderIssue(issue *keymapv1.ValidationIssue) string {
	var content string
	switch v := issue.Issue.(type) {
	case *keymapv1.ValidationIssue_KeybindConflict:
		c := v.KeybindConflict
		var actionLines []string
		for _, action := range c.Actions {
			if action.EditorCommand != "" {
				actionLines = append(actionLines, fmt.Sprintf("%s (%s)", action.EditorCommand, action.Action))
			} else {
				actionLines = append(actionLines, action.Action)
			}
		}
		content = fmt.Sprintf("Keybind Conflict: %s is mapped to multiple actions:\n  - %s",
			keyStyle.Render(c.Keybinding),
			actionStyle.Render(strings.Join(actionLines, "\n  - ")))
	case *keymapv1.ValidationIssue_DanglingAction:
		d := v.DanglingAction
		suggestion := ""
		if d.Suggestion != "" {
			suggestion = fmt.Sprintf(" (Did you mean %s?)", actionStyle.Render(d.Suggestion))
		}
		content = fmt.Sprintf("Dangling Action: %s for key %s does not exist.%s",
			actionStyle.Render(d.Action), keyStyle.Render(d.Keybinding), suggestion)
	case *keymapv1.ValidationIssue_UnsupportedAction:
		u := v.UnsupportedAction
		content = fmt.Sprintf("Unsupported Action: %s (on key %s) is not supported for target %s.",
			actionStyle.Render(u.Action), keyStyle.Render(u.Keybinding), keyStyle.Render(u.TargetEditor))
	case *keymapv1.ValidationIssue_DuplicateMapping:
		d := v.DuplicateMapping
		content = fmt.Sprintf("Duplicate Mapping: Action %s with key %s is defined multiple times.",
			actionStyle.Render(d.Action), keyStyle.Render(d.Keybinding))
	case *keymapv1.ValidationIssue_PotentialShadowing:
		p := v.PotentialShadowing
		content = fmt.Sprintf("Potential Shadowing: Key %s (for action %s) might override a system or editor default on %s.",
			keyStyle.Render(p.Keybinding), actionStyle.Render(p.Action), keyStyle.Render(p.TargetEditor))
	default:
		content = "Unknown issue type."
	}
	return issueStyle.Render(content) + "\n"
}
