package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	columnWidthActionName = 48
	columnWidthKeybinding = 30
	viewHeightMargin      = 7
	minViewHeight         = 6
)

var _ tea.Model = (*KeymapViewModel)(nil)

// KeymapViewModel is a read-only TUI model to present current OneKeymapSetting.
type KeymapViewModel struct {
	setting *keymapv1.Keymap
	mc      *mappings.MappingConfig

	// derived state
	categories       []string
	selectedCategory int

	rows  []table.Row
	table table.Model

	width, height int
}

func NewKeymapViewModel(setting *keymapv1.Keymap, mc *mappings.MappingConfig) tea.Model {
	m := &KeymapViewModel{
		setting: setting,
		mc:      mc,
	}
	m.initCategories()
	m.rebuildRows()
	m.table = table.New(
		table.WithColumns(
			[]table.Column{
				{Title: "Action", Width: columnWidthActionName},
				{Title: "Keybinding", Width: columnWidthKeybinding},
			},
		),
		table.WithRows(m.rows),
		table.WithFocused(true),
	)
	return m
}

func (m *KeymapViewModel) initCategories() {
	catSet := map[string]struct{}{}
	for _, kb := range m.setting.GetActions() {
		if mapping := m.mc.FindByUniversalAction(kb.GetName()); mapping != nil {
			if mapping.Category != "" {
				catSet[mapping.Category] = struct{}{}
			}
		}
	}
	cats := make([]string, 0, len(catSet))
	for c := range catSet {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	m.categories = append([]string{"All"}, cats...)
}

func (m *KeymapViewModel) Init() tea.Cmd { return nil }

func (m *KeymapViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "left":
			if m.selectedCategory > 0 {
				m.selectedCategory--
				m.rebuildRows()
				m.table.SetRows(m.rows)
			}
		case "right":
			if m.selectedCategory < len(m.categories)-1 {
				m.selectedCategory++
				m.rebuildRows()
				m.table.SetRows(m.rows)
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// leave 5 lines for headers/help/detail; adjust table height
		h := m.height - viewHeightMargin
		if h < minViewHeight {
			h = minViewHeight
		}
		m.table.SetHeight(h)
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *KeymapViewModel) View() string {
	var b strings.Builder
	// Header title
	b.WriteString("OneKeymap Viewer (read-only)  —  Use ←/→ switch category, ↑/↓ navigate, q to quit\n")
	// Categories
	b.WriteString("Categories: ")
	for i, c := range m.categories {
		if i == m.selectedCategory {
			b.WriteString(fmt.Sprintf("[ %s ] ", c))
		} else {
			b.WriteString(fmt.Sprintf("  %s   ", c))
		}
	}
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n")

	// Detail of selected row
	selectedID := m.selectedActionID()
	if selectedID != "" {
		m.appendSelectedActionDetails(&b, selectedID)
	}
	return b.String()
}

func (m *KeymapViewModel) appendSelectedActionDetails(b *strings.Builder, selectedID string) {
	b.WriteString("\n")
	fmt.Fprintf(b, "Action: %s\n", selectedID)
	if mapping := m.mc.FindByUniversalAction(selectedID); mapping != nil {
		if mapping.Name != "" {
			fmt.Fprintf(b, "Description: %s\n", mapping.Name)
		} else if mapping.Description != "" {
			fmt.Fprintf(b, "Description: %s\n", mapping.Description)
		}
		if mapping.Category != "" {
			fmt.Fprintf(b, "Category: %s\n", mapping.Category)
		}
	}
}

func (m *KeymapViewModel) selectedActionID() string {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return row[0]
}

func (m *KeymapViewModel) rebuildRows() {
	// Aggregate keybindings by action id
	agg := map[string][]string{}
	for _, ab := range m.setting.GetActions() {
		if !m.includeByCategory(ab.GetName()) {
			continue
		}
		for _, b := range ab.GetBindings() {
			k := keybindingToString(b)
			if k == "" {
				continue
			}
			agg[ab.GetName()] = append(agg[ab.GetName()], k)
		}
	}
	ids := make([]string, 0, len(agg))
	for id := range agg {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	rows := make([]table.Row, 0, len(ids))
	for _, id := range ids {
		keys := dedup(agg[id])

		mapping := m.mc.FindByUniversalAction(id)
		d := func() string {
			if mapping == nil {
				return id
			}
			return mapping.Name
		}()
		rows = append(rows, table.Row{d, strings.Join(keys, " or ")})
	}
	m.rows = rows
}

func (m *KeymapViewModel) includeByCategory(actionID string) bool {
	if m.selectedCategory == 0 { // All
		return true
	}
	cat := m.categories[m.selectedCategory]
	if mapping := m.mc.FindByUniversalAction(actionID); mapping != nil {
		return mapping.Category == cat
	}
	return false
}

func dedup(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func keybindingToString(kb *keymapv1.KeybindingReadable) string {
	if kb == nil {
		return ""
	}
	f, err := keymap.NewKeyBinding(kb).Format(platform.PlatformMacOS, "+")
	if err != nil {
		return ""
	}
	return f
}
