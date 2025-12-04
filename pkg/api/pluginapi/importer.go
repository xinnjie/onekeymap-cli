package pluginapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
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
}

type ImportSkipReport struct {
	SkipActions []ImportSkipAction
}

type ImportSkipAction struct {
	// EditorSpecificAction is the action name in the editor-specific format
	EditorSpecificAction string
	Error                error
}
