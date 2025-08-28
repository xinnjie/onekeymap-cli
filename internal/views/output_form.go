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
	_ tea.Model = (*outputFormModel)(nil)
)

// outputFormModel collects inputs for exporting onekeymap.json to a target editor.
type outputFormModel struct {
	form *huh.Form

	pluginRegistry *plugins.Registry

	needSelectEditor bool
	needInput        bool
	needOutput       bool

	Editor                   *string
	OnekeymapConfigInput     *string
	EditorKeymapConfigOutput *string

	OnekeymapConfigPlaceHolder string
}

func NewOutputFormModel(
	registry *plugins.Registry,
	needSelectEditor, needInput, needOutput bool,
	editor, onekeymapConfigInput, editorKeymapConfigOutput *string,
	onekeymapConfigPlaceHolder string,
) (*outputFormModel, error) {
	m := &outputFormModel{
		pluginRegistry:             registry,
		needSelectEditor:           needSelectEditor,
		needInput:                  needInput,
		needOutput:                 needOutput,
		Editor:                     editor,
		OnekeymapConfigInput:       onekeymapConfigInput,
		EditorKeymapConfigOutput:   editorKeymapConfigOutput,
		OnekeymapConfigPlaceHolder: onekeymapConfigPlaceHolder,
	}
	if err := m.build(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *outputFormModel) build() error {
	if !m.needSelectEditor && !m.needInput && !m.needOutput {
		return fmt.Errorf("form not needed")
	}

	var groups []*huh.Group

	if m.needSelectEditor {
		names := m.GetExporterNames()
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
					Title("Select target editor").
					Options(opts...).
					Value(m.Editor),
			),
		)
	}

	if m.needInput {
		placeholderInput := m.OnekeymapConfigPlaceHolder
		groups = append(groups,
			huh.NewGroup(
				huh.NewInput().
					Key("input").
					Title("Source onekeymap.json path").
					Placeholder(placeholderInput).
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
					Value(m.OnekeymapConfigInput),
			),
		)
	}

	if m.needOutput {
		placeholderOutput := ""
		groups = append(groups,
			huh.NewGroup(
				huh.NewInput().
					Key("output").
					TitleFunc(func() string {
						return "Output config path for " + *m.Editor
					}, &m.Editor).
					PlaceholderFunc(func() string {
						if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
							if v, err := p.DefaultConfigPath(); err == nil {
								placeholderOutput = v[0]
							}
						}
						return placeholderOutput
					}, &m.Editor).
					Value(m.EditorKeymapConfigOutput),
			),
		)
	}

	m.form = huh.NewForm(groups...)
	return nil
}

func (m *outputFormModel) GetExporterNames() []string {
	exporterNames := make([]string, 0)
	for _, name := range m.pluginRegistry.GetNames() {
		plugin, ok := m.pluginRegistry.Get(pluginapi.EditorType(name))
		if !ok {
			continue
		}

		_, err := plugin.Exporter()
		if err != nil {
			continue
		}
		exporterNames = append(exporterNames, name)
	}
	return exporterNames
}

func (m *outputFormModel) Init() tea.Cmd { return m.form.Init() }

func (m *outputFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}

	if m.form.State == huh.StateCompleted {
		// Apply placeholders if user left fields empty
		if m.needInput && *m.OnekeymapConfigInput == "" {
			*m.OnekeymapConfigInput = m.OnekeymapConfigPlaceHolder
		}
		if m.needOutput && *m.EditorKeymapConfigOutput == "" {
			if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
				if v, err := p.DefaultConfigPath(); err == nil {
					*m.EditorKeymapConfigOutput = v[0]
				}
			}
		}
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

func (m *outputFormModel) View() string {
	return m.form.View()
}
