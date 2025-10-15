package views

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	columnWidthActionName = 48
	columnWidthKeybinding = 30
	minViewHeight         = 6
	categoryPanelWidth    = 25
	viewHeightMargin      = 7
)

var _ tea.Model = (*KeymapViewModel)(nil)

// KeymapViewModel is a read-only TUI model to present current OneKeymapSetting.
type KeymapViewModel struct {
	setting *keymapv1.Keymap
	mc      *mappings.MappingConfig

	// category selection
	categories       []string
	selectedCategory string
	categorySelect   *huh.Select[string]
	actionTable      table.Model

	width, height int
}

func NewKeymapViewModel(setting *keymapv1.Keymap, mc *mappings.MappingConfig) tea.Model {
	m := &KeymapViewModel{
		setting: setting,
		mc:      mc,
	}
	m.initCategories()
	m.selectedCategory = "All"

	// Create huh.Select for categories
	options := make([]huh.Option[string], len(m.categories))
	for i, cat := range m.categories {
		options[i] = huh.NewOption(cat, cat)
	}
	m.categorySelect = huh.NewSelect[string]().
		Title("Categories").
		Options(options...).
		Value(&m.selectedCategory)

	m.actionTable = table.New(
		table.WithColumns(
			[]table.Column{
				{Title: "Action", Width: columnWidthActionName},
				{Title: "Keybinding", Width: columnWidthKeybinding},
			},
		),
		table.WithRows(m.keybindingRows()),
		table.WithFocused(true),
	)
	return m
}

func (m *KeymapViewModel) Init() tea.Cmd {
	return m.categorySelect.Init()
}

func (m *KeymapViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// nolint:goconst // key strings for TUI input are clearer inline here
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "tab":
			// Move to next category
			m.moveToNextCategory(1)
			m.actionTable.SetRows(m.keybindingRows())
		case "shift+tab":
			// Move to previous category
			m.moveToNextCategory(-1)
			m.actionTable.SetRows(m.keybindingRows())
		}
	case tea.WindowSizeMsg:
		h := msg.Height - viewHeightMargin
		if h < minViewHeight {
			h = minViewHeight
		}
		m.width, m.height = msg.Width, h
		m.actionTable.SetHeight(h)
	}

	var tableCmd tea.Cmd
	m.actionTable, tableCmd = m.actionTable.Update(msg)
	cmds = append(cmds, tableCmd)

	return m, tea.Batch(cmds...)
}

func (m *KeymapViewModel) View() string {
	help := "OneKeymap Viewer (read-only)  —  Use ↑/↓ navigate, Tab/Shift+Tab to switch category, q to quit\n\n"

	// Left panel: category select
	categoryPanel := lipgloss.NewStyle().
		Width(categoryPanelWidth).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(m.categorySelect.View())

	// Right panel: table and details
	rightPanelContent := m.actionTable.View() + "\n"
	selectedID := m.selectedActionID()
	if selectedID != "" {
		details := newActionDetailsViewModel(selectedID, m.mc)
		rightPanelContent += details.View()
	}

	rightPanel := lipgloss.NewStyle().
		Width(m.width - categoryPanelWidth).
		Render(rightPanelContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, categoryPanel, rightPanel)
	body = lipgloss.JoinVertical(lipgloss.Top, body, help)

	return body
}

func (m *KeymapViewModel) selectedActionID() string {
	row := m.actionTable.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return row[0]
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

func (m *KeymapViewModel) moveToNextCategory(delta int) {
	if len(m.categories) == 0 {
		return
	}

	// Find current index
	currentIdx := 0
	for i, cat := range m.categories {
		if cat == m.selectedCategory {
			currentIdx = i
			break
		}
	}

	// Calculate next index with wrapping
	nextIdx := (currentIdx + delta) % len(m.categories)
	if nextIdx < 0 {
		nextIdx += len(m.categories)
	}

	m.selectedCategory = m.categories[nextIdx]
	m.categorySelect.Value(&m.selectedCategory)
}

func (m *KeymapViewModel) includeAction(actionID string) bool {
	if m.selectedCategory == "All" {
		return true
	}
	if mapping := m.mc.FindByUniversalAction(actionID); mapping != nil {
		return mapping.Category == m.selectedCategory
	}
	return false
}

func (m *KeymapViewModel) keybindingRows() []table.Row {
	// Aggregate keybindings by action id
	agg := map[string][]string{}
	for _, ab := range m.setting.GetActions() {
		if !m.includeAction(ab.GetName()) {
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
	return rows
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
