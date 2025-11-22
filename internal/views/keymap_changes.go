package views

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	keybindingpkg "github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

const (
	columnWidthType    = 8
	columnWidthAction  = 50
	tableDefaultHeight = 18
	tableHeightMargin  = 4
	minTableHeight     = 6
	defaultColumnWidth = 20
)

var (
	_ tea.Model = (*keymapChangesModel)(nil)
)

// TODO(xinnjie): low priority, choose which change not to apply
type keymapChangesModel struct {
	table   table.Model
	changes *importerapi.KeymapChanges

	confirm    *bool
	confirming bool
	form       *huh.Form
}

func NewKeymapChangesModel(changes *importerapi.KeymapChanges, confirm *bool) tea.Model {
	// Compute dynamic widths for Before/After based on actual content
	beforeW, afterW := measureBeforeAfterWidths(changes)
	cols := []table.Column{
		{Title: "Type", Width: columnWidthType},
		{Title: "Action", Width: columnWidthAction},
		{Title: "Before", Width: beforeW},
		{Title: "After", Width: afterW},
	}
	var rows []table.Row
	if changes != nil {
		for _, kb := range changes.Remove {
			rows = append(rows, table.Row{"Remove", redMinus(kb.Name), redMinus(formatKeyBinding(&kb)), ""})
		}
		for _, kb := range changes.Add {
			rows = append(rows, table.Row{"Add", greenPlus(kb.Name), "", greenPlus(formatKeyBinding(&kb))})
		}
		for _, diff := range changes.Update {
			action := ""
			if diff.Before.Name != "" {
				action = diff.Before.Name
			} else if diff.After.Name != "" {
				action = diff.After.Name
			}
			rows = append(
				rows,
				table.Row{
					"Update",
					action,
					redMinus(formatKeyBinding(&diff.Before)),
					greenPlus(formatKeyBinding(&diff.After)),
				},
			)
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableDefaultHeight),
		table.WithFocused(true),
	)
	return keymapChangesModel{table: t, changes: changes, confirm: confirm}
}

func (m keymapChangesModel) Init() tea.Cmd { return nil }

func (m keymapChangesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.confirming {
		return m.handleConfirmingState(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
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
		h := msg.Height - tableHeightMargin
		if h < minTableHeight {
			h = minTableHeight
		}
		if h > len(m.table.Rows())+1 {
			h = len(m.table.Rows()) + 1
		}
		m.table.SetHeight(h)
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m keymapChangesModel) handleConfirmingState(msg tea.Msg) (tea.Model, tea.Cmd) {
	// When confirming, delegate to the form
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyRunes:
			if km.String() == "q" {
				if m.confirm != nil {
					*m.confirm = false
				}
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.confirm != nil {
				*m.confirm = false
			}
			return m, tea.Quit
		}
	}
	var form tea.Model
	var cmd tea.Cmd
	form, cmd = m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		return m, tea.Quit
	}
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

func formatKeyBinding(action *keymap.Action) string {
	if action == nil {
		return ""
	}
	var parts []string
	for _, b := range action.Bindings {
		f := b.String(keybindingpkg.FormatOption{
			Platform:  platform.PlatformMacOS,
			Separator: "+",
		})
		if f == "" {
			continue
		}
		parts = append(parts, f)
	}
	return strings.Join(parts, " or ")
}

// measureBeforeAfterWidths calculates the suitable column widths for Before/After
// using the display length of formatted keybindings, with sensible caps.
func measureBeforeAfterWidths(changes *importerapi.KeymapChanges) (before, after int) {
	const (
		minW      = 12
		maxW      = 80
		prefixLen = 2 // account for "+ " or "- " prefixes
	)
	if changes == nil {
		return defaultColumnWidth, defaultColumnWidth
	}
	maxBefore := 0
	maxAfter := 0
	for _, kb := range changes.Remove {
		l := utf8.RuneCountInString(formatKeyBinding(&kb)) + prefixLen
		if l > maxBefore {
			maxBefore = l
		}
	}
	for _, kb := range changes.Add {
		l := utf8.RuneCountInString(formatKeyBinding(&kb)) + prefixLen
		if l > maxAfter {
			maxAfter = l
		}
	}
	for _, diff := range changes.Update {
		if diff.Before.Name != "" {
			l := utf8.RuneCountInString(formatKeyBinding(&diff.Before)) + prefixLen
			if l > maxBefore {
				maxBefore = l
			}
		}
		if diff.After.Name != "" {
			l := utf8.RuneCountInString(formatKeyBinding(&diff.After)) + prefixLen
			if l > maxAfter {
				maxAfter = l
			}
		}
	}
	if maxBefore < minW {
		maxBefore = minW
	}
	if maxAfter < minW {
		maxAfter = minW
	}
	if maxBefore > maxW {
		maxBefore = maxW
	}
	if maxAfter > maxW {
		maxAfter = maxW
	}
	return maxBefore, maxAfter
}
