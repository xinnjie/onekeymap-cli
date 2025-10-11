package views

import (
	"errors"
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

var (
	_ tea.Model = (*ExportFormModel)(nil)
)

// ExportFormModel collects inputs for exporting onekeymap.json to a target editor.
type ExportFormModel struct {
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
) (*ExportFormModel, error) {
	m := &ExportFormModel{
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

func (m *ExportFormModel) build() error {
	if !m.needSelectEditor && !m.needInput && !m.needOutput {
		return errors.New("form not needed")
	}

	var groups []*huh.Group

	if m.needSelectEditor {
		editorOpts := m.getExporterOptions()
		finalOpts := buildEditorSelectOptions(editorOpts)

		if len(finalOpts) == 0 {
			return errors.New("no editor plugins available")
		}
		groups = append(groups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("editor").
					Title("Select target editor").
					Options(finalOpts...).
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
							if v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
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

func (m *ExportFormModel) getExporterOptions() []editorSelectorOption {
	var options []editorSelectorOption
	for _, name := range m.pluginRegistry.GetNames() {
		editorType := pluginapi.EditorType(name)
		plugin, ok := m.pluginRegistry.Get(editorType)
		if !ok {
			continue
		}

		_, err := plugin.Exporter()
		if err != nil {
			continue
		}

		_, installed, _ := plugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
		options = append(options, editorSelectorOption{
			displayName: editorType.AppName(),
			editorType:  name,
			installed:   installed,
		})
	}

	// Sort by display name for consistent order
	sort.Slice(options, func(i, j int) bool {
		return options[i].displayName < options[j].displayName
	})

	return options
}

func (m *ExportFormModel) Init() tea.Cmd { return m.form.Init() }

func (m *ExportFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		// nolint:goconst // key strings for TUI input are clearer inline here
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}

	if m.form.State == huh.StateCompleted {
		m.applyPlaceholders()
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

func (m *ExportFormModel) View() string {
	return m.form.View()
}

func (m *ExportFormModel) applyPlaceholders() {
	// Apply placeholders if user left fields empty
	if m.needInput && *m.OnekeymapConfigInput == "" {
		*m.OnekeymapConfigInput = m.OnekeymapConfigPlaceHolder
	}
	if m.needOutput && *m.EditorKeymapConfigOutput == "" {
		if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
			if v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
				*m.EditorKeymapConfigOutput = v[0]
			}
		}
	}
}
