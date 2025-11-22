package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
)

type ValidationReportModel struct {
	report   *validateapi.ValidationReport
	viewport viewport.Model
	ready    bool
}

func NewValidationReportModel(report *validateapi.ValidationReport) ValidationReportModel {
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
		switch msg.Type {
		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		case tea.KeyEsc, tea.KeyEnter, tea.KeyCtrlC:
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

	s := m.report.Summary
	b.WriteString(summaryStyle.Render(
		fmt.Sprintf("Source: %s | Mappings Processed: %d | Succeeded: %d",
			m.report.SourceEditor, s.MappingsProcessed, s.MappingsSucceeded),
	))
	b.WriteString("\n")

	if len(m.report.Issues) > 0 {
		b.WriteString(errorHeaderStyle.Render(fmt.Sprintf("Issues (%d)", len(m.report.Issues))))
		for _, issue := range m.report.Issues {
			b.WriteString(renderIssue(issue))
		}
	}

	if len(m.report.Warnings) > 0 {
		b.WriteString(warningHeaderStyle.Render(fmt.Sprintf("Warnings (%d)", len(m.report.Warnings))))
		for _, warning := range m.report.Warnings {
			b.WriteString(renderIssue(warning))
		}
	}

	return b.String()
}

func renderIssue(issue validateapi.ValidationIssue) string {
	var content string
	switch issue.Type {
	case validateapi.IssueTypeKeybindConflict:
		if c, ok := issue.Details.(validateapi.KeybindConflict); ok {
			var actionLines []string
			for _, action := range c.Actions {
				if action.Context != "" {
					actionLines = append(actionLines, fmt.Sprintf("%s (%s)", action.Context, action.ActionID))
				} else {
					actionLines = append(actionLines, action.ActionID)
				}
			}
			content = fmt.Sprintf("Keybind Conflict: %s is mapped to multiple actions:\n  - %s",
				keyStyle.Render(c.Keybinding),
				actionStyle.Render(strings.Join(actionLines, "\n  - ")))
		}
	case validateapi.IssueTypeDanglingAction:
		if d, ok := issue.Details.(validateapi.DanglingAction); ok {
			suggestion := ""
			if d.Suggestion != "" {
				suggestion = fmt.Sprintf(" (%s)", actionStyle.Render(d.Suggestion))
			}
			content = fmt.Sprintf("Dangling Action: %s does not exist in target %s.%s",
				actionStyle.Render(d.Action), keyStyle.Render(d.TargetEditor), suggestion)
		}
	case validateapi.IssueTypeUnsupportedAction:
		if u, ok := issue.Details.(validateapi.UnsupportedAction); ok {
			content = fmt.Sprintf("Unsupported Action: %s (on key %s) is not supported for target %s.",
				actionStyle.Render(u.Action), keyStyle.Render(u.Keybinding), keyStyle.Render(u.TargetEditor))
		}
	case validateapi.IssueTypeDuplicateMapping:
		if d, ok := issue.Details.(validateapi.DuplicateMapping); ok {
			content = fmt.Sprintf("Duplicate Mapping: Action %s with key %s is defined multiple times.",
				actionStyle.Render(d.Action), keyStyle.Render(d.Keybinding))
		}
	case validateapi.IssueTypePotentialShadowing:
		if p, ok := issue.Details.(validateapi.PotentialShadowing); ok {
			content = fmt.Sprintf("Potential Shadowing: Key %s (for action %s). %s",
				keyStyle.Render(p.Keybinding), actionStyle.Render(p.Action), p.CriticalShortcutDescription)
		}
	default:
		content = "Unknown issue type."
	}
	return issueStyle.Render(content) + "\n"
}
