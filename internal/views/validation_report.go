package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
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
		//nolint:goconst // key strings for TUI input are clearer inline here
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

	s := m.report.GetSummary()
	b.WriteString(summaryStyle.Render(
		fmt.Sprintf("Source: %s | Mappings Processed: %d | Succeeded: %d",
			m.report.GetSourceEditor(), s.GetMappingsProcessed(), s.GetMappingsSucceeded()),
	))
	b.WriteString("\n")

	if len(m.report.GetIssues()) > 0 {
		b.WriteString(errorHeaderStyle.Render(fmt.Sprintf("Issues (%d)", len(m.report.GetIssues()))))
		for _, issue := range m.report.GetIssues() {
			b.WriteString(renderIssue(issue))
		}
	}

	if len(m.report.GetWarnings()) > 0 {
		b.WriteString(warningHeaderStyle.Render(fmt.Sprintf("Warnings (%d)", len(m.report.GetWarnings()))))
		for _, warning := range m.report.GetWarnings() {
			b.WriteString(renderIssue(warning))
		}
	}

	return b.String()
}

func renderIssue(issue *keymapv1.ValidationIssue) string {
	var content string
	switch v := issue.GetIssue().(type) {
	case *keymapv1.ValidationIssue_KeybindConflict:
		c := v.KeybindConflict
		var actionLines []string
		for _, action := range c.GetActions() {
			if action.GetEditorCommand() != "" {
				actionLines = append(actionLines, fmt.Sprintf("%s (%s)", action.GetEditorCommand(), action.GetAction()))
			} else {
				actionLines = append(actionLines, action.GetAction())
			}
		}
		content = fmt.Sprintf("Keybind Conflict: %s is mapped to multiple actions:\n  - %s",
			keyStyle.Render(c.GetKeybinding()),
			actionStyle.Render(strings.Join(actionLines, "\n  - ")))
	case *keymapv1.ValidationIssue_DanglingAction:
		d := v.DanglingAction
		suggestion := ""
		if d.GetSuggestion() != "" {
			suggestion = fmt.Sprintf(" (Did you mean %s?)", actionStyle.Render(d.GetSuggestion()))
		}
		content = fmt.Sprintf("Dangling Action: %s for key %s does not exist.%s",
			actionStyle.Render(d.GetAction()), keyStyle.Render(d.GetKeybinding()), suggestion)
	case *keymapv1.ValidationIssue_UnsupportedAction:
		u := v.UnsupportedAction
		content = fmt.Sprintf("Unsupported Action: %s (on key %s) is not supported for target %s.",
			actionStyle.Render(u.GetAction()), keyStyle.Render(u.GetKeybinding()), keyStyle.Render(u.GetTargetEditor()))
	case *keymapv1.ValidationIssue_DuplicateMapping:
		d := v.DuplicateMapping
		content = fmt.Sprintf("Duplicate Mapping: Action %s with key %s is defined multiple times.",
			actionStyle.Render(d.GetAction()), keyStyle.Render(d.GetKeybinding()))
	case *keymapv1.ValidationIssue_PotentialShadowing:
		p := v.PotentialShadowing
		content = fmt.Sprintf("Potential Shadowing: Key %s (for action %s) might override a system or editor default on %s.",
			keyStyle.Render(p.GetKeybinding()), actionStyle.Render(p.GetAction()), keyStyle.Render(p.GetTargetEditor()))
	default:
		content = "Unknown issue type."
	}
	return issueStyle.Render(content) + "\n"
}
