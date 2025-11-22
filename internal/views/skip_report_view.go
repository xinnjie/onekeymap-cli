package views

import (
	"errors"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	actionColumnWidth     = 40
	keybindingColumnWidth = 20
	reasonColumnWidth     = 50
	tableHeight           = 15
	keyQuit               = "ctrl+c"
)

type SkipReportViewModel struct {
	title string
	table table.Model
	style lipgloss.Style
	help  string
}

func NewSkipReportViewModel(skipActions []pluginapi.ExportSkipAction) SkipReportViewModel {
	columns := []table.Column{
		{Title: "Action", Width: actionColumnWidth},
		{Title: "Keybinding", Width: keybindingColumnWidth},
		{Title: "Reason", Width: reasonColumnWidth},
	}

	rows := make([]table.Row, len(skipActions))
	for i, sk := range skipActions {
		var keybindingStr string
		var e *pluginapi.EditorSupportOnlyOneKeybindingPerActionError
		if errors.As(sk.Error, &e) && e.SkipKeybinding != nil {
			binding := keymap.NewKeyBinding(&keymapv1.KeybindingReadable{KeyChords: e.SkipKeybinding})
			keybindingStr, _ = binding.Format(platform.PlatformMacOS, "+")
		}

		rows[i] = table.Row{
			sk.Action,
			keybindingStr,
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

	t.SetStyles(s)

	return SkipReportViewModel{
		title: "Skipped Actions for Export",
		table: t,
		style: baseStyle,
		help:  "Press q or ctrl+c to quit",
	}
}

func (m SkipReportViewModel) Init() tea.Cmd {
	return nil
}

func (m SkipReportViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyRunes:
			if keyMsg.String() == "q" {
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEnter:
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m SkipReportViewModel) View() string {
	return lipgloss.NewStyle().Bold(true).PaddingBottom(1).Render(m.title) + "\n" +
		m.style.Render(m.table.View()) + "\n" +
		helpStyle.Render(m.help) + "\n" +
		helpStyle.Render(m.table.HelpView())
}
