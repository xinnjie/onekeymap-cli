package views

import (
	"errors"
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

type editorOption struct {
	name      string
	installed bool
}

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

func NewImportFormModel(
	registry *plugins.Registry,
	needSelectEditor, needInput, needOutput bool,
	editor, editorKeymapConfigInput, onekeymapConfigOutput *string,
	onekeymapConfigPlaceHolder string,
) (*importFormModel, error) {
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
		return errors.New("form not needed")
	}

	var groups []*huh.Group

	if m.needSelectEditor {
		editorOpts := m.getImporterOptions()

		installed := make([]huh.Option[string], 0)
		uninstalled := make([]huh.Option[string], 0)

		for _, opt := range editorOpts {
			label := opt.name
			if !opt.installed {
				label = fmt.Sprintf("%s (uninstalled)", opt.name)
			}
			huhOpt := huh.NewOption(label, opt.name)
			if opt.installed {
				installed = append(installed, huhOpt)
			} else {
				uninstalled = append(uninstalled, huhOpt)
			}
		}

		finalOpts := make([]huh.Option[string], 0, len(installed)+len(uninstalled))
		finalOpts = append(finalOpts, installed...)
		finalOpts = append(finalOpts, uninstalled...)

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
							if v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
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

func (m *importFormModel) getImporterOptions() []editorOption {
	var options []editorOption
	for _, name := range m.pluginRegistry.GetNames() {
		plugin, ok := m.pluginRegistry.Get(pluginapi.EditorType(name))
		if !ok {
			continue
		}

		if _, err := plugin.Importer(); err != nil {
			continue
		}

		_, installed, _ := plugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
		options = append(options, editorOption{name: name, installed: installed})
	}

	// Sort by name for consistent order
	sort.Slice(options, func(i, j int) bool {
		return options[i].name < options[j].name
	})

	return options
}

// tea.Model minimal implementations (not used directly, kept for future extension).
func (m *importFormModel) Init() tea.Cmd { return m.form.Init() }

func (m *importFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}
	if m.form.State == huh.StateCompleted {
		// Fill placeholder if user input is empty
		if m.needInput && *m.EditorKeymapConfigInput == "" {
			if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.Editor)); ok {
				if v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
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
