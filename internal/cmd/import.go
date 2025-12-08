package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
)

type importFlags struct {
	from        string
	input       string
	output      string
	interactive bool
	backup      bool
}

//nolint:dupl // Import/Export command constructors are intentionally symmetrical; limited duplication keeps each isolated and clearer
func NewCmdImport() *cobra.Command {
	f := importFlags{}
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an editor's keymap to the universal format",
		RunE: importRun(&f, func() (*slog.Logger, *registry.Registry, importerapi.Importer) {
			return cmdLogger, cmdPluginRegistry, cmdImportService
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.from, "from", "", "Source editor to import from (e.g., vscode, zed)")
	cmd.Flags().StringVar(&f.output, "output", "", "Path to save the generated onekeymap.json file")
	cmd.Flags().
		StringVar(&f.input, "input", "", "Optional: Path to the source editor's config file (overrides env vars)")
	cmd.Flags().BoolVar(&f.interactive, "interactive", true, "Run in interactive mode")
	cmd.Flags().BoolVar(&f.backup, "backup", false, "Create a backup of the target editor's keymap")

	// Add completion for 'from' flag
	_ = cmd.RegisterFlagCompletionFunc(
		"from",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return cmdPluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
		},
	)

	return cmd
}

func importRun(
	f *importFlags,
	dependencies func() (*slog.Logger, *registry.Registry, importerapi.Importer),
) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		logger, pluginRegistry, importService := dependencies()
		onekeymapConfig := viper.GetString("onekeymap")

		if f.interactive {
			return importRunInteractive(cmd, f, logger, pluginRegistry, importService, onekeymapConfig)
		}

		return importRunNonInteractive(cmd, f, logger, pluginRegistry, importService, onekeymapConfig)
	}
}

func importRunInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	pluginRegistry *registry.Registry,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	if err := prepareInteractiveImportFlags(cmd, f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	return executeImportInteractive(cmd, f, logger, importService, onekeymapConfig)
}

func importRunNonInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	pluginRegistry *registry.Registry,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	if err := prepareNonInteractiveImportFlags(f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	return executeImportNonInteractive(cmd, f, logger, importService, onekeymapConfig)
}

func executeImportInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	var (
		file io.ReadCloser
		err  error
	)

	if f.input != "" {
		if f.from == "basekeymap" {
			// For basekeymap, f.input is the base keymap name, not a file path
			file = io.NopCloser(strings.NewReader(f.input))
		} else {
			file, err = os.Open(f.input)
			if err != nil {
				logger.Error("Failed to open input file", "path", f.input, "error", err)
				return err
			}
			defer func() { _ = file.Close() }()
		}
	}

	baseConfig := loadBaseConfig(f.output, onekeymapConfig, logger)

	opts := importerapi.ImportOptions{
		EditorType:  pluginapi.EditorType(f.from),
		InputStream: file,
		Base:        baseConfig,
	}

	result, err := importService.Import(cmd.Context(), opts)
	if err != nil {
		logger.Error("import failed", "error", err)
		return err
	}

	if result != nil {
		printImportSummary(cmd, result)
	}

	if result.Changes.HasChanges() {
		confirmed, err := runImportChangesPreview(result.Changes)
		if err != nil {
			logger.Warn("failed to render changes preview", "error", err)
		}
		if !confirmed {
			logger.Info("User cancelled applying changes; no file will be written")
			return nil
		}
	} else {
		cmd.Println("No changes to import - file will not be updated")
	}

	return saveImportResult(f.output, result, logger)
}

func executeImportNonInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	var (
		file io.ReadCloser
		err  error
	)

	if f.input != "" {
		if f.from == "basekeymap" {
			// For basekeymap, f.input is the base keymap name, not a file path
			file = io.NopCloser(strings.NewReader(f.input))
		} else {
			file, err = os.Open(f.input)
			if err != nil {
				logger.Error("Failed to open input file", "path", f.input, "error", err)
				return err
			}
			defer func() { _ = file.Close() }()
		}
	}

	baseConfig := loadBaseConfig(f.output, onekeymapConfig, logger)

	opts := importerapi.ImportOptions{
		EditorType:  pluginapi.EditorType(f.from),
		InputStream: file,
		Base:        baseConfig,
	}

	result, err := importService.Import(cmd.Context(), opts)
	if err != nil {
		logger.Error("import failed", "error", err)
		return err
	}

	if result != nil {
		printImportSummary(cmd, result)
	}

	return saveImportResult(f.output, result, logger)
}

func loadBaseConfig(outputPath, onekeymapConfig string, logger *slog.Logger) keymap.Keymap {
	basePath := outputPath
	if basePath == "" {
		basePath = onekeymapConfig
	}
	if basePath == "" {
		return keymap.Keymap{}
	}

	baseConfigFile, err := os.Open(basePath)
	if err != nil {
		logger.Debug("Base config file not found, skip loading base config", "path", basePath)
		return keymap.Keymap{}
	}
	defer func() { _ = baseConfigFile.Close() }()

	cfg, lerr := keymap.Load(baseConfigFile, keymap.LoadOptions{})
	if lerr != nil {
		logger.Warn("Failed to load base keymap, treat as no base config", "error", lerr)
		return keymap.Keymap{}
	}

	return cfg
}

func saveImportResult(outputPath string, result *importerapi.ImportResult, logger *slog.Logger) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
		logger.Error("Failed to create output directory", "dir", filepath.Dir(outputPath), "error", err)
		return err
	}
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		logger.Error("Failed to create output file", "error", err)
		return err
	}
	defer func() {
		_ = outputFile.Close()
	}()

	if result == nil || len(result.Setting.Actions) == 0 {
		logger.Warn("No keymaps imported; nothing to save")
		return nil
	}

	// Use new API Save
	saveOpt := keymap.SaveOptions{Platform: platform.PlatformMacOS}
	if err := keymap.Save(outputFile, result.Setting, saveOpt); err != nil {
		logger.Error("Failed to save config file", "error", err)
		return err
	}

	logger.Info("Successfully imported keymap", "output", outputPath)
	if result.Report != nil {
		logger.Debug("Import report", "report", result.Report)
	}
	return nil
}

type importSkippedView struct {
	Action          string
	KeybindingCount int
	Reason          string
}

type importSummaryView struct {
	TotalImported     int
	Skipped           []importSkippedView
	HasValidation     bool
	ValidationSource  string
	MappingsProcessed int
	MappingsSucceeded int
	Issues            []string
	Warnings          []string
}

// nolint: gochecknoglobals
var importSummaryTemplate = template.Must(template.New("importSummary").Parse(`
Import Summary:
  ✓ {{ .TotalImported }} actions imported into onekeymap
{{- $skippedCount := len .Skipped }}
{{- if eq $skippedCount 0 }}
  ✗ 0 editor actions skipped
{{- else }}
  ✗ {{ $skippedCount }} editor actions skipped:
{{- range .Skipped }}
    - {{ .Action }}{{ if gt .KeybindingCount 0 }} ({{ .KeybindingCount }} keybindings){{ end }}{{ if .Reason }}: {{ .Reason }}{{ end }}
{{- end }}
{{- end }}
{{- if .HasValidation }}

Validation Summary:
  Source: {{ .ValidationSource }} | Mappings Processed: {{ .MappingsProcessed }} | Succeeded: {{ .MappingsSucceeded }}
  Issues: {{ len .Issues }}, Warnings: {{ len .Warnings }}
{{- if gt (len .Issues) 0 }}
  Issues ({{ len .Issues }}):
{{- range .Issues }}
    - {{ . }}
{{- end }}
{{- end }}
{{- if gt (len .Warnings) 0 }}
  Warnings ({{ len .Warnings }}):
{{- range .Warnings }}
    - {{ . }}
{{- end }}
{{- end }}
{{- end }}
`))

func printImportSummary(cmd *cobra.Command, result *importerapi.ImportResult) {
	view := importSummaryView{
		TotalImported: len(result.Setting.Actions),
	}

	for _, sk := range result.SkipReport.SkipActions {
		item := importSkippedView{
			Action:          sk.EditorSpecificAction,
			KeybindingCount: len(sk.Keybindings),
		}
		if sk.Error != nil {
			item.Reason = sk.Error.Error()
		}
		view.Skipped = append(view.Skipped, item)
	}

	if result.Report != nil {
		rep := result.Report
		view.HasValidation = true
		view.ValidationSource = rep.SourceEditor
		view.MappingsProcessed = rep.Summary.MappingsProcessed
		view.MappingsSucceeded = rep.Summary.MappingsSucceeded
		for _, issue := range rep.Issues {
			view.Issues = append(view.Issues, renderValidationIssueInline(issue))
		}
		for _, warning := range rep.Warnings {
			view.Warnings = append(view.Warnings, renderValidationIssueInline(warning))
		}
	}

	var buf bytes.Buffer
	if err := importSummaryTemplate.Execute(&buf, view); err != nil {
		// Fallback to minimal output if template rendering fails
		cmd.Println()
		cmd.Println("Import Summary:")
		cmd.Printf("  ✓ %d actions imported into onekeymap\n", view.TotalImported)
		return
	}

	cmd.Println()
	cmd.Print(buf.String())
}

// renderValidationIssueInline renders a single validation issue in a compact textual form,
// mirroring the semantics of views.renderIssue but without TUI styling.
func renderValidationIssueInline(issue validateapi.ValidationIssue) string {
	switch issue.Type {
	case validateapi.IssueTypeKeybindConflict:
		if c, ok := issue.Details.(validateapi.KeybindConflict); ok {
			var actionLines []string
			for _, action := range c.Actions {
				if action.Context != "" {
					actionLines = append(actionLines, fmt.Sprintf("%s (%s)", action.Context, action.ActionID))
				} else {
					actionLines = append(actionLines, action.ActionID)
				}
			}
			return fmt.Sprintf(
				"Keybind Conflict: %s is mapped to multiple actions:\n      - %s",
				c.Keybinding,
				strings.Join(actionLines, "\n      - "),
			)
		}
	case validateapi.IssueTypeDanglingAction:
		if d, ok := issue.Details.(validateapi.DanglingAction); ok {
			suggestion := ""
			if d.Suggestion != "" {
				suggestion = fmt.Sprintf(" (%s)", d.Suggestion)
			}
			return fmt.Sprintf(
				"Dangling Action: %s does not exist in target %s.%s",
				d.Action,
				d.TargetEditor,
				suggestion,
			)
		}
	case validateapi.IssueTypeUnsupportedAction:
		if u, ok := issue.Details.(validateapi.UnsupportedAction); ok {
			return fmt.Sprintf(
				"Unsupported Action: %s (on key %s) is not supported for target %s.",
				u.Action,
				u.Keybinding,
				u.TargetEditor,
			)
		}
	case validateapi.IssueTypeDuplicateMapping:
		if d, ok := issue.Details.(validateapi.DuplicateMapping); ok {
			return fmt.Sprintf(
				"Duplicate Mapping: Action %s with key %s is defined multiple times.",
				d.Action,
				d.Keybinding,
			)
		}
	case validateapi.IssueTypePotentialShadowing:
		if p, ok := issue.Details.(validateapi.PotentialShadowing); ok {
			return fmt.Sprintf(
				"Potential Shadowing: Key %s (for action %s). %s",
				p.Keybinding,
				p.Action,
				p.CriticalShortcutDescription,
			)
		}
	}

	return "Unknown issue type."
}

func handleInteractiveImportFlags(
	cmd *cobra.Command,
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *registry.Registry,
	logger *slog.Logger,
) error {
	needSelectEditor := !cmd.Flags().Changed("from") || f.from == ""
	needInput := !cmd.Flags().Changed("input") || f.input == ""
	needOutput := !cmd.Flags().Changed("output") || f.output == ""

	if needSelectEditor || needInput || needOutput {
		if err := runImportForm(
			pluginRegistry,
			&f.from,
			&f.input,
			&f.output,
			onekeymapConfig,
			needSelectEditor,
			needInput,
			needOutput,
		); err != nil {
			return err
		}
	}

	logger.DebugContext(cmd.Context(), "after input propmt", "from", f.from, "input", f.input, "output", f.output)
	return nil
}

func prepareInteractiveImportFlags(
	cmd *cobra.Command,
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *registry.Registry,
	logger *slog.Logger,
) error {
	if err := handleInteractiveImportFlags(cmd, f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	p, ok := pluginRegistry.Get(pluginapi.EditorType(f.from))
	if !ok {
		logger.Error("Editor not found", "editor", f.from)
		return fmt.Errorf("editor %s not found", f.from)
	}

	if f.input == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", f.from))
		if configPath != "" {
			f.input = configPath
			logger.Info("Using keymap path from config", "editor", f.from, "path", configPath)
		} else {
			v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}
	}
	return nil
}

func prepareNonInteractiveImportFlags(
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *registry.Registry,
	logger *slog.Logger,
) error {
	if f.from == "" {
		return errors.New("flag --from is required")
	}
	if onekeymapConfig != "" {
		f.output = onekeymapConfig
	}

	p, ok := pluginRegistry.Get(pluginapi.EditorType(f.from))
	if !ok {
		logger.Error("Editor not found", "editor", f.from)
		return fmt.Errorf("editor %s not found", f.from)
	}

	if f.input == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", f.from))
		if configPath != "" {
			f.input = configPath
			logger.Info("Using keymap path from config", "editor", f.from, "path", configPath)
		} else {
			v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}
	}
	return nil
}

func runImportChangesPreview(changes *importerapi.KeymapChanges) (bool, error) {
	confirmed := true
	m := views.NewKeymapChangesModel(changes, &confirmed)
	_, err := tea.NewProgram(m).Run()
	return confirmed, err
}

// runImportForm runs the interactive import form and returns the selected values.
// All TUI logic is encapsulated here to keep cmd/import.go simple.
func runImportForm(
	pluginRegistry *registry.Registry,
	from, input, output *string,
	onekeymapConfigPlaceHolder string,
	needSelectEditor, needInput, needOutput bool,
) error {
	m, err := views.NewImportFormModel(
		pluginRegistry,
		needSelectEditor,
		needInput,
		needOutput,
		from,
		input,
		output,
		onekeymapConfigPlaceHolder,
	)
	if err != nil {
		return err
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		if errors.Is(err, tea.ErrProgramKilled) {
			os.Exit(0)
		}
		return err
	}
	return nil
}
