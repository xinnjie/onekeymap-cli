package views

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	columnWidthActionName = 48
	columnWidthKeybinding = 30
	minViewHeight         = 10 // Need space for header + rows
	categoryPanelWidth    = 25
	viewHeightMargin      = 10 // Space for help text + details
	detailsHeight         = 8  // Estimated height for action details
	defaultInitialWidth   = 120
	defaultInitialHeight  = 30
	tablePaddingWidth     = 6 // Padding for borders and spacing
	panelBorderWidth      = 4 // Width adjustment for panel borders
)

var _ tea.Model = (*KeymapViewModel)(nil)

// FileChangedMsg is sent when the watched file changes
type FileChangedMsg struct {
	Path string
}

// FileErrorMsg is sent when there's an error watching the file
type FileErrorMsg struct {
	Err error
}

// KeymapViewModel is a read-only TUI model to present current OneKeymapSetting.
type KeymapViewModel struct {
	setting  *keymapv1.Keymap
	mc       *mappings.MappingConfig
	filePath string // Path to the onekeymap.json file being watched

	// category selection
	categories       []string
	selectedCategory string
	categorySelect   *huh.Select[string]
	actionTable      table.Model

	width, height int
	errorMsg      string // Error message to display
}

func NewKeymapViewModel(setting *keymapv1.Keymap, mc *mappings.MappingConfig, filePath string) tea.Model {
	m := &KeymapViewModel{
		setting:  setting,
		mc:       mc,
		filePath: filePath,
		width:    defaultInitialWidth,
		height:   defaultInitialHeight,
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

	// Calculate initial table dimensions
	initialTableWidth := m.width - categoryPanelWidth - tablePaddingWidth
	initialTableHeight := m.height - viewHeightMargin - detailsHeight
	if initialTableHeight < minViewHeight {
		initialTableHeight = minViewHeight
	}

	m.actionTable = table.New(
		table.WithColumns(
			[]table.Column{
				{Title: "Action", Width: columnWidthActionName},
				{Title: "Keybinding", Width: columnWidthKeybinding},
			},
		),
		table.WithRows(m.keybindingRows()),
		table.WithFocused(true),
		table.WithHeight(initialTableHeight),
		table.WithWidth(initialTableWidth),
	)
	return m
}

func (m *KeymapViewModel) Init() tea.Cmd {
	return m.categorySelect.Init()
}

func (m *KeymapViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case FileChangedMsg:
		// File changed, reload the keymap
		m.errorMsg = "" // Clear any previous error
		if err := m.reloadKeymap(); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to reload keymap: %v", err)
		} else {
			// Reinitialize categories and update table
			m.initCategories()
			m.actionTable.SetRows(m.keybindingRows())
		}
		return m, nil

	case FileErrorMsg:
		// Error watching file
		m.errorMsg = fmt.Sprintf("File watch error: %v", msg.Err)
		return m, nil

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
		m.width, m.height = msg.Width, msg.Height

		// Calculate available width for table (total - category panel)
		tableWidth := m.width - categoryPanelWidth - tablePaddingWidth

		// Calculate available height for table (total - help - details - margins)
		tableHeight := m.height - viewHeightMargin - detailsHeight
		if tableHeight < minViewHeight {
			tableHeight = minViewHeight
		}

		// Update table dimensions
		m.actionTable.SetHeight(tableHeight)
		m.actionTable.SetWidth(tableWidth)
	}

	var tableCmd tea.Cmd
	m.actionTable, tableCmd = m.actionTable.Update(msg)
	cmds = append(cmds, tableCmd)

	return m, tea.Batch(cmds...)
}

func (m *KeymapViewModel) View() string {
	// Help text at the top
	helpText := "OneKeymap Viewer (read-only)  —  Use ↑/↓ navigate, Tab/Shift+Tab to switch category, q to quit"
	if m.errorMsg != "" {
		helpText = fmt.Sprintf("⚠ %s", m.errorMsg)
	}
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(helpText)

	// Left panel: category select
	categoryPanel := lipgloss.NewStyle().
		Width(categoryPanelWidth).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(m.categorySelect.View())

	// Right panel: table and details
	var rightPanelContent string
	if m.width > 0 && m.height > 0 {
		rightPanelContent = m.actionTable.View()
		selectedID := m.selectedActionID()
		if selectedID != "" {
			details := newActionDetailsViewModel(selectedID, m.mc)
			rightPanelContent += "\n" + details.View()
		}
	}

	rightPanel := lipgloss.NewStyle().
		Width(m.width - categoryPanelWidth - panelBorderWidth).
		Render(rightPanelContent)

	// Combine panels horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, categoryPanel, rightPanel)

	// Combine help and main content vertically
	return lipgloss.JoinVertical(lipgloss.Left, help, "", mainContent)
}

func (m *KeymapViewModel) selectedActionID() string {
	row := m.actionTable.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	for _, actionName := range m.mc.Mappings {
		if actionName.Name == row[0] {
			return actionName.ID
		}
	}
	return ""
}

func (m *KeymapViewModel) initCategories() {
	catSet := map[string]struct{}{}
	for _, kb := range m.setting.GetActions() {
		if mapping := m.mc.Get(kb.GetName()); mapping != nil {
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
	if mapping := m.mc.Get(actionID); mapping != nil {
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
			k := b.GetKeyChordsReadable()
			if k == "" {
				continue
			}
			agg[ab.GetName()] = append(agg[ab.GetName()], k)
		}
	}

	return m.sortDeterministic(agg)
}

func (m *KeymapViewModel) sortDeterministic(action2keybindings map[string][]string) []table.Row {
	ids := make([]string, 0, len(action2keybindings))
	for id := range action2keybindings {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	rows := make([]table.Row, 0, len(ids))
	for _, id := range ids {
		keys := dedup(action2keybindings[id])

		mapping := m.mc.Get(id)
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

// reloadKeymap reloads the keymap from the file
func (m *KeymapViewModel) reloadKeymap() error {
	file, err := os.Open(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	setting, err := keymap.Load(file)
	if err != nil {
		return fmt.Errorf("failed to parse keymap: %w", err)
	}

	m.setting = setting
	return nil
}
