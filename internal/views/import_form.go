package views

import (
	"errors"
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

var (
	_ tea.Model = (*ImportFormModel)(nil)
)

const selectOtherOption = "Other (manual input)"

// ImportFormModel represents the import form UI model.
type ImportFormModel struct {
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

func NewImportFormModel(
	registry *plugins.Registry,
	needSelectEditor, needInput, needOutput bool,
	editor, editorKeymapConfigInput, onekeymapConfigOutput *string,
	onekeymapConfigPlaceHolder string,
) (*ImportFormModel, error) {
	m := &ImportFormModel{
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

func (m *ImportFormModel) build() error {
	if !m.needSelectEditor && !m.needInput && !m.needOutput {
		return errors.New("form not needed")
	}

	var groups []*huh.Group

	if m.needSelectEditor {
		editorOpts := m.getImporterOptions()
		finalOpts := buildEditorSelectOptions(editorOpts)

		if len(finalOpts) == 0 {
			return errors.New("no editor plugins available")
		}

		groups = append(groups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("editor").
					Title("Select source editor").
					Options(finalOpts...).
					Value(m.Editor),
			),
		)
	}

	if m.needInput {
		groups = append(groups, m.createInputGroup()...)
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

func (m *ImportFormModel) getImporterOptions() []editorSelectorOption {
	var options []editorSelectorOption
	for _, name := range m.pluginRegistry.GetNames() {
		editorType := pluginapi.EditorType(name)
		plugin, ok := m.pluginRegistry.Get(editorType)
		if !ok {
			continue
		}

		if _, err := plugin.Importer(); err != nil {
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

// Init initializes the import form model.
func (m *ImportFormModel) Init() tea.Cmd { return m.form.Init() }

func (m *ImportFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		// nolint:goconst // key strings for TUI input are clearer inline here
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}
	if m.form.State == huh.StateCompleted {
		m.fillPlaceholders()
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

func (m *ImportFormModel) View() string {
	return m.form.View()
}

func (m *ImportFormModel) fillPlaceholders() {
	// Fill placeholder if user input is empty
	if m.needInput && *m.EditorKeymapConfigInput == "" {
		configs := m.configPaths()
		if len(configs) > 0 {
			*m.EditorKeymapConfigInput = configs[0]
		}
	}
	if m.needOutput && *m.OnekeymapConfigOutput == "" {
		*m.OnekeymapConfigOutput = m.OnekeymapConfigPlaceHolder
	}
}

func (m *ImportFormModel) createInputGroup() []*huh.Group {
	multipleConfigAvailableSelect := huh.NewGroup(
		huh.NewSelect[string]().
			TitleFunc(func() string {
				return "Select config path for " + *m.Editor
			}, &m.Editor).
			OptionsFunc(func() []huh.Option[string] {
				configs := m.configPaths()
				var opts []huh.Option[string]
				for _, path := range configs {
					opts = append(opts, huh.NewOption(path, path))
				}
				opts = append(opts, huh.NewOption(selectOtherOption, selectOtherOption))
				return opts
			}, &m.Editor).
			Value(m.EditorKeymapConfigInput),
	).WithHideFunc(func() bool {
		configs := m.configPaths()
		return len(configs) <= 1
	})

	multipleConfigAvailableInput := huh.NewGroup(
		huh.NewInput().
			TitleFunc(func() string {
				return "Input config path for " + *m.Editor
			}, &m.Editor).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("config path is required")
				}
				if _, err := os.Stat(s); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", s)
				}
				return nil
			}).
			Value(m.EditorKeymapConfigInput),
	).WithHideFunc(func() bool {
		configs := m.configPaths()
		return len(configs) <= 1 || *m.EditorKeymapConfigInput != selectOtherOption
	})

	lessConfigAvailableInput := huh.NewGroup(
		huh.NewInput().
			TitleFunc(func() string {
				return "Input config path for " + *m.Editor
			}, &m.Editor).
			PlaceholderFunc(func() string {
				return m.placeholder()
			}, &m.Editor).
			// Ensure the input file exists
			Validate(func(s string) error {
				filePath := s
				if filePath == "" {
					filePath = m.placeholder()
				}
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", filePath)
				}
				return nil
			}).
			Value(m.EditorKeymapConfigInput),
	).WithHideFunc(func() bool {
		configs := m.configPaths()
		return len(configs) > 1
	})

	return []*huh.Group{
		multipleConfigAvailableSelect,
		multipleConfigAvailableInput,
		lessConfigAvailableInput,
	}
}

func (m *ImportFormModel) configPaths() []string {
	if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
		if paths, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
			return paths
		}
	}
	return []string{}
}

func (m *ImportFormModel) placeholder() string {
	configs := m.configPaths()
	if len(configs) > 0 {
		return configs[0]
	}
	return ""
}
