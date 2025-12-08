package pluginapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	keybinding "github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

type PluginImportOption struct {
	// SourcePlatform specifies the platform of the source keybindings being imported.
	// This affects how modifier keys are parsed (e.g., "cmd" for macOS, "win" for Windows).
	// If empty, defaults to the current runtime platform.
	SourcePlatform platform.Platform
}

type PluginImporter interface {
	// Import reads the editor's configuration source and converts it into the
	// universal onekeymap KeymapSetting format.
	Import(ctx context.Context, source io.Reader, opts PluginImportOption) (PluginImportResult, error)
}

// ConfigDetectOptions provides configuration for a config detect operation.
type ConfigDetectOptions struct {
	// Whether to in sandbox mode, in sandbox mode, shell command lookup is not effective, like `code` command for vscode can not be found
	Sandbox bool
}

// PluginImportResult contains the result of an import operation.
type PluginImportResult struct {
	Keymap keymap.Keymap

	Report PluginImportReport
}

// PluginImportReport contains details about the import operation.
type PluginImportReport struct {
	// SkipReport contains details about actions that were skipped during the import operation.
	SkipReport ImportSkipReport

	// ImportedReport contains detailed import results for each editor keybinding.
	// This is used by the importer service to compute ImportCoverage.
	ImportedReport ImportedReport
}

type ImportSkipReport struct {
	SkipActions []ImportSkipAction
}

type ImportSkipAction struct {
	// EditorSpecificAction is the action name in the editor-specific format
	EditorSpecificAction string
	// Keybindings are the keybindings found in the editor config.
	Keybindings []keybinding.Keybinding
	// Reason why this action is skipped during importing
	Error error
}

// ImportedReport contains detailed import results for editor keybindings.
type ImportedReport struct {
	// Results contains import outcome for each editor keybinding group
	Results []KeybindingImportResult
}

// KeybindingImportResult describes the import outcome for editor keybindings.
type KeybindingImportResult struct {
	// MappedAction is the universal action this keybinding was mapped to.
	// e.g., "actions.editor.copy"
	MappedAction string

	// EditorSpecificAction is the original editor-specific command.
	// e.g., "editor.action.clipboardCopyAction" (VSCode)
	EditorSpecificAction string

	// OriginalKeybindings are the keybindings found in the editor config.
	OriginalKeybindings []keybinding.Keybinding

	// ImportedKeybindings are the keybindings that were actually imported.
	// May differ from Original due editor constraint, e.g. Xcode can only configure one keybinding per action.
	ImportedKeybindings []keybinding.Keybinding

	// Reason explains any discrepancy (optional)
	Reason string
}
