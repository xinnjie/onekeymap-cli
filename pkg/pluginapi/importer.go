package pluginapi

import (
	"context"
	"io"

	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type PluginImportOption struct {
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
	Keymap *keymapv1.Keymap

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
