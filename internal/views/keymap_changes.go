package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/charmbracelet/bubbles/table"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var (
	_ tea.Model = (*keymapChangesModel)(nil)
)

// TODO(xinnjie): low priority, choose which change not to apply
type keymapChangesModel struct {
	table   table.Model
	changes *importapi.KeymapChanges

	confirm    *bool
	confirming bool
	form       *huh.Form
}

func NewKeymapChangesModel(changes *importapi.KeymapChanges, confirm *bool) tea.Model {
	cols := []table.Column{
		{Title: "Type", Width: 8},
		{Title: "Action", Width: 50},
		{Title: "Before", Width: 20},
		{Title: "After", Width: 20},
	}
	var rows []table.Row
	if changes != nil {
		for _, kb := range changes.Remove {
			rows = append(rows, table.Row{"Remove", redMinus(kb.GetId()), redMinus(formatKeyBinding(kb)), ""})
		}
		for _, kb := range changes.Add {
			rows = append(rows, table.Row{"Add", greenPlus(kb.GetId()), "", greenPlus(formatKeyBinding(kb))})
		}
		for _, diff := range changes.Update {
			action := ""
			if diff.Before != nil {
				action = diff.Before.GetId()
			} else if diff.After != nil {
				action = diff.After.GetId()
			}
			rows = append(rows, table.Row{"Update", action, redMinus(formatKeyBinding(diff.Before)), greenPlus(formatKeyBinding(diff.After))})
		}
	}

	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithHeight(18), table.WithFocused(true))
	return keymapChangesModel{table: t, changes: changes, confirm: confirm}
}

func (m keymapChangesModel) Init() tea.Cmd { return nil }

func (m keymapChangesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.confirming {
		// When confirming, delegate to the form
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc", "q":
				if m.confirm != nil {
					*m.confirm = false
				}
				return m, tea.Quit
			}
		}
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			return m, tea.Quit
		}
		return m, cmd
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "enter":
			// Show confirm dialog
			m.confirming = true
			c := huh.NewConfirm().
				Title("Apply keymap changes?").
				Description(fmt.Sprintf("%d to add, %d to change, %d to remove", len(m.changes.Add), len(m.changes.Update), len(m.changes.Remove))).
				Affirmative("Apply").
				Negative("Cancel").
				Value(m.confirm)
			m.form = huh.NewForm(huh.NewGroup(c))
			return m, m.form.Init()
		}
	case tea.WindowSizeMsg:
		h := msg.Height - 4
		if h < 6 {
			h = 6
		}
		if h > len(m.table.Rows())+1 {
			h = len(m.table.Rows()) + 1
		}
		m.table.SetHeight(h)
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m keymapChangesModel) View() string {
	var b strings.Builder
	b.WriteString("Keymap Import Changes Preview (press enter to review and confirm, press ctrl+c or q to quit)\n\n")
	b.WriteString(m.table.View())
	b.WriteString("\n")

	if m.confirming && m.form != nil {
		b.WriteString(m.form.View())
		b.WriteString("\n")
	}
	return b.String()
}

func greenPlus(s string) string {
	return fmt.Sprintf("\x1b[32m+\x1b[0m %s", s)
}

func redMinus(s string) string {
	return fmt.Sprintf("\x1b[31m-\x1b[0m %s", s)
}

func formatKeyBinding(kb *keymapv1.KeyBinding) string {
	if kb == nil {
		return ""
	}
	f, err := keymap.NewKeyBinding(kb).Format(platform.PlatformMacOS, "+")
	if err != nil {
		return ""
	}
	return f
}
