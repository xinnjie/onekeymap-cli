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

type CategoryViewModel struct {
	categories []string
	selected   int
	mc         *mappings.MappingConfig
}

func newCategoryViewModel(setting *keymapv1.Keymap, mc *mappings.MappingConfig) CategoryViewModel {
	cv := CategoryViewModel{
		mc: mc,
	}
	cv.initCategories(setting)
	return cv
}

func (cv *CategoryViewModel) initCategories(setting *keymapv1.Keymap) {
	catSet := map[string]struct{}{}
	for _, kb := range setting.GetActions() {
		if mapping := cv.mc.FindByUniversalAction(kb.GetName()); mapping != nil {
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
	cv.categories = append([]string{"All"}, cats...)
}

func (cv *CategoryViewModel) SelectPrev() bool {
	if cv.selected > 0 {
		cv.selected--
		return true
	}
	return false
}

func (cv *CategoryViewModel) SelectNext() bool {
	if cv.selected < len(cv.categories)-1 {
		cv.selected++
		return true
	}
	return false
}

func (cv *CategoryViewModel) Include(actionID string) bool {
	if cv.selected == 0 {
		return true
	}
	cat := cv.categories[cv.selected]
	if mapping := cv.mc.FindByUniversalAction(actionID); mapping != nil {
		return mapping.Category == cat
	}
	return false
}

func (cv *CategoryViewModel) View() string {
	var b strings.Builder
	for i, c := range cv.categories {
		if i == cv.selected {
			b.WriteString(fmt.Sprintf("[ %s ] ", c))
			continue
		}
		b.WriteString(fmt.Sprintf("  %s   ", c))
	}
	return b.String()
}

// KeymapViewModel is a read-only TUI model to present current OneKeymapSetting.
type KeymapViewModel struct {
	setting *keymapv1.Keymap
	mc      *mappings.MappingConfig

	// derived state
	category CategoryViewModel

	rows  []table.Row
	table table.Model

	width, height int
}

func NewKeymapViewModel(setting *keymapv1.Keymap, mc *mappings.MappingConfig) tea.Model {
	m := &KeymapViewModel{
		setting: setting,
		mc:      mc,
	}
	m.category = newCategoryViewModel(setting, mc)
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

func (m *KeymapViewModel) Init() tea.Cmd { return nil }

func (m *KeymapViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// nolint:goconst // key strings for TUI input are clearer inline here
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "left":
			if m.category.SelectPrev() {
				m.rebuildRows()
				m.table.SetRows(m.rows)
			}
		case "right":
			if m.category.SelectNext() {
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
	b.WriteString(m.category.View())
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n")

	// Detail of selected row
	selectedID := m.selectedActionID()
	if selectedID != "" {
		details := newActionDetailsViewModel(selectedID, m.mc)
		b.WriteString(details.View())
	}
	return b.String()
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
		if !m.category.Include(ab.GetName()) {
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
