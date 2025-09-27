package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var (
	_ tea.Model = (*MigrateFormModel)(nil)
)

type MigrateFormModel struct {
	form *huh.Form

	pluginRegistry *plugins.Registry
	EditorFrom     *string
	EditorTo       *string
	Input          *string
	Output         *string

	// placeholders
	fromPlaceholder string
	toPlaceholder   string
}

func NewMigrateFormModel(registry *plugins.Registry, from, to, input, output *string) *MigrateFormModel {
	m := &MigrateFormModel{
		pluginRegistry: registry,
		EditorFrom:     from,
		EditorTo:       to,
		Input:          input,
		Output:         output,
	}
	m.buildForm()
	return m
}

func (m *MigrateFormModel) buildForm() {
	editorNames := m.pluginRegistry.GetNames()
	var editorOptions []string
	for _, name := range editorNames {
		editorOptions = append(editorOptions, string(pluginapi.EditorType(name)))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("from").
				Title("From which editor?").
				Options(huh.NewOptions(editorOptions...)...).
				Value(m.EditorFrom),

			huh.NewInput().
				Key("input").
				Title("Source config path").
				Description("The path to the source editor's keymap config file.").
				PlaceholderFunc(func() string {
					if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.EditorFrom)); ok {
						if paths, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil &&
							len(paths) > 0 {
							m.fromPlaceholder = paths[0]
							return paths[0]
						}
					}
					return ""
				}, m.EditorFrom).
				Value(m.Input),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("to").
				Title("To which editor?").
				Options(huh.NewOptions(editorOptions...)...).
				Value(m.EditorTo),

			huh.NewInput().
				Key("output").
				Title("Target config path").
				Description("The path to the target editor's keymap config file.").
				PlaceholderFunc(func() string {
					if p, ok := m.pluginRegistry.Get(pluginapi.EditorType(*m.EditorTo)); ok {
						if paths, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil &&
							len(paths) > 0 {
							m.toPlaceholder = paths[0]
							return paths[0]
						}
					}
					return ""
				}, m.EditorTo).
				Value(m.Output),
		),
	)
}

func (m *MigrateFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MigrateFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Interrupt
		}
	}

	if m.form.State == huh.StateCompleted {
		if *m.Input == "" {
			*m.Input = m.fromPlaceholder
		}
		if *m.Output == "" {
			*m.Output = m.toPlaceholder
		}
		return m, tea.Quit
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

func (m *MigrateFormModel) View() string {
	return m.form.View()
}
