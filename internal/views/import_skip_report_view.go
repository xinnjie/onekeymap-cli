package views

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type ImportSkipReportViewModel struct {
	title string
	table table.Model
	style lipgloss.Style
	help  string
}

func NewImportSkipReportViewModel(report pluginapi.ImportSkipReport) ImportSkipReportViewModel {
	columns := []table.Column{
		{Title: "Editor Action", Width: actionColumnWidth},
		{Title: "Reason", Width: actionColumnWidth + reasonColumnWidth - keybindingColumnWidth},
	}

	rows := make([]table.Row, len(report.SkipActions))
	for i, sk := range report.SkipActions {
		rows[i] = table.Row{
			sk.EditorSpecificAction,
			sk.Error.Error(),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	t.SetStyles(s)

	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return ImportSkipReportViewModel{
		title: "Skipped Actions for Import",
		table: t,
		style: baseStyle,
		help:  "Press q or ctrl+c to quit",
	}
}

func (m ImportSkipReportViewModel) Init() tea.Cmd {
	return nil
}

func (m ImportSkipReportViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", tea.KeyCtrlC.String(), tea.KeyEnter.String():
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ImportSkipReportViewModel) View() string {
	return lipgloss.NewStyle().Bold(true).PaddingBottom(1).Render(m.title) + "\n" +
		m.style.Render(m.table.View()) + "\n" +
		helpStyle.Render(m.help) + "\n" +
		helpStyle.Render(m.table.HelpView())
}
