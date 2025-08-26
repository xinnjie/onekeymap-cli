package views

import (
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var (
	_ tea.Model = (*importFormModel)(nil)
)

type importFormModel struct {
	form *huh.Form

	pluginRegistry *plugins.Registry

	needSelectEditor           bool
	needInput                  bool
	needOutput                 bool
	Editor                     *string
	EditorKeymapConfigInput    *string
	OnekeymapConfigOutput      *string
	OnekeymapConfigPlaceHolder string
}

func NewImportFormModel(registry *plugins.Registry,
	needSelectEditor, needInput, needOutput bool,
	editor, editorKeymapConfigInput, onekeymapConfigOutput *string, onekeymapConfigPlaceHolder string) (*importFormModel, error) {
	m := &importFormModel{
		pluginRegistry:             registry,
		needSelectEditor:           needSelectEditor,
		needInput:                  needInput,
		needOutput:                 needOutput,
		Editor:                     editor,
		EditorKeymapConfigInput:    editorKeymapConfigInput,
		OnekeymapConfigOutput:      onekeymapConfigOutput,
		OnekeymapConfigPlaceHolder: onekeymapConfigPlaceHolder,
	}
	if err := m.build(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *importFormModel) build() error {
	if !m.needSelectEditor && !m.needInput && !m.needOutput {
		return fmt.Errorf("form not needed")
	}

	var groups []*huh.Group

	if m.needSelectEditor {
		names := m.GetImporterNames()
		sort.Strings(names)
		opts := make([]huh.Option[string], 0, len(names))
		for _, n := range names {
			opts = append(opts, huh.NewOption(n, n))
		}
		if len(opts) == 0 {
			return fmt.Errorf("no editor plugins available")
		}
		groups = append(groups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("editor").
					Title("Select source editor").
					Options(opts...).
					Value(m.Editor),
			),
		)
	}

	if m.needInput {
		placeholderInput := ""
		groups = append(groups,
			huh.NewGroup(
				huh.NewInput().
					Key("input").
					TitleFunc(func() string {
						return "Input config path for " + *m.Editor
					}, &m.Editor).
					PlaceholderFunc(func() string {
						if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
							if v, err := p.DefaultConfigPath(); err == nil {
								placeholderInput = v[0]
							}
						}
						return placeholderInput
					}, &m.Editor).
					// Ensure the input file exists
					Validate(func(s string) error {
						filePath := s
						if filePath == "" {
							filePath = placeholderInput
						}
						if _, err := os.Stat(filePath); os.IsNotExist(err) {
							return fmt.Errorf("file does not exist: %s", filePath)
						}
						return nil
					}).
					Value(m.EditorKeymapConfigInput),
			),
		)
	}

	if m.needOutput {
		groups = append(groups,
			huh.NewGroup(
				huh.NewInput().
					Key("output").
					Title("Output file path").
					Placeholder(m.OnekeymapConfigPlaceHolder).
					Value(m.OnekeymapConfigOutput),
			),
		)
	}

	m.form = huh.NewForm(groups...)
	return nil
}

func (m *importFormModel) GetImporterNames() []string {
	importerNames := make([]string, 0)
	for _, name := range m.pluginRegistry.GetNames() {
		plugin, ok := m.pluginRegistry.Get(pluginapi.EditorType(name))
		if !ok {
			continue
		}

		_, err := plugin.Importer()
		if err != nil {
			continue
		}
		importerNames = append(importerNames, name)
	}
	return importerNames
}

// tea.Model minimal implementations (not used directly, kept for future extension)
func (m *importFormModel) Init() tea.Cmd { return m.form.Init() }

func (m *importFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}
	if m.form.State == huh.StateCompleted {
		// Fill placeholder if user input is empty
		if m.needInput && *m.EditorKeymapConfigInput == "" {
			if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
				if v, err := p.DefaultConfigPath(); err == nil {
					*m.EditorKeymapConfigInput = v[0]
				}
			}
		}
		if m.needOutput && *m.OnekeymapConfigOutput == "" {
			*m.OnekeymapConfigOutput = m.OnekeymapConfigPlaceHolder
		}
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

func (m *importFormModel) View() string {
	return m.form.View()
}
