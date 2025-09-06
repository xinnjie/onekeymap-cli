package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var _ tea.Model = (*KeymapViewModel)(nil)

// KeymapViewModel is a read-only TUI model to present current OneKeymapSetting.
type KeymapViewModel struct {
	setting *keymapv1.KeymapSetting
	mc      *mappings.MappingConfig

	// derived state
	categories       []string
	selectedCategory int

	rows  []table.Row
	table table.Model

	width, height int
}

func NewKeymapViewModel(setting *keymapv1.KeymapSetting, mc *mappings.MappingConfig) tea.Model {
	m := &KeymapViewModel{
		setting: setting,
		mc:      mc,
	}
	m.initCategories()
	m.rebuildRows()
	m.table = table.New(
		table.WithColumns([]table.Column{{Title: "Action", Width: 48}, {Title: "Keybinding", Width: 30}}),
		table.WithRows(m.rows),
		table.WithFocused(true),
	)
	return m
}

func (m *KeymapViewModel) initCategories() {
	catSet := map[string]struct{}{}
	for _, kb := range m.setting.GetKeybindings() {
		if mapping := m.mc.FindByUniversalAction(kb.GetId()); mapping != nil {
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
		h := m.height - 7
		if h < 6 {
			h = 6
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
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Action: %s\n", selectedID))
		if mapping := m.mc.FindByUniversalAction(selectedID); mapping != nil {
			if mapping.Name != "" {
				b.WriteString(fmt.Sprintf("Description: %s\n", mapping.Name))
			} else if mapping.Description != "" {
				b.WriteString(fmt.Sprintf("Description: %s\n", mapping.Description))
			}
			if mapping.Category != "" {
				b.WriteString(fmt.Sprintf("Category: %s\n", mapping.Category))
			}
		}
	}
	return b.String()
}

func (m *KeymapViewModel) selectedActionID() string {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return fmt.Sprint(row[0])
}

func (m *KeymapViewModel) rebuildRows() {
	// Aggregate keybindings by action id
	agg := map[string][]string{}
	for _, kb := range m.setting.GetKeybindings() {
		if !m.includeByCategory(kb.GetId()) {
			continue
		}
		k := keybindingToString(kb)
		if k == "" {
			continue
		}
		agg[kb.GetId()] = append(agg[kb.GetId()], k)
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

func keybindingToString(kb *keymapv1.KeyBinding) string {
	if kb == nil {
		return ""
	}
	f, err := keymap.NewKeyBinding(kb).Format(platform.PlatformMacOS, "+")
	if err != nil {
		return ""
	}
	return f
}
