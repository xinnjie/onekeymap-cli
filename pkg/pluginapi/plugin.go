package pluginapi

import (
	"context"
	"io"

	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// Plugin is the core interface that all editor plugins must implement.
// It defines the contract for importing and exporting keymaps.
type Plugin interface {
	// EditorType returns the unique identifier for the plugin (e.g., "vscode", "zed").
	EditorType() EditorType

	// ConfigDetect returns the default path to the editor's configuration file based on the platform.
	// Return multiple paths if the editor has multiple configuration files.
	ConfigDetect(opts ConfigDetectOptions) (paths []string, installed bool, err error)

	// Importer returns an instance of PluginImporter for the plugin.
	// Return ErrNotSupported if the plugin does not support importing.
	Importer() (PluginImporter, error)

	// Exporter returns an instance of PluginExporter for the plugin.
	// Return ErrNotSupported if the plugin does not support exporting.
	Exporter() (PluginExporter, error)
}

type PluginImportOption struct {
}

type PluginImporter interface {
	// Import reads the editor's configuration source and converts it into the
	// universal onekeymap KeymapSetting format.
	Import(ctx context.Context, source io.Reader, opts PluginImportOption) (*keymapv1.Keymap, error)
}

type PluginExporter interface {
	// Export takes a universal KeymapSetting and writes it to an editor-specific
	// configuration destination.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting *keymapv1.Keymap,
		opts PluginExportOption,
	) (*PluginExportReport, error)
}

// PluginExportOption provides configuration for an export operation.
type PluginExportOption struct {
	// ExistingConfig provides a reader for the editor's current keymap
	// configuration. If provided, the exporter will merge the settings
	// instead of overwriting the entire file, preserving any keybindings
	// not managed by onekeymap. If nil, the export will be destructive,
	// creating a new file from scratch.
	ExistingConfig io.Reader
}

// PluginExportReport details issues encountered during an export operation.
type PluginExportReport struct {
	// The diff between the base and the exported keymap.
	// If base is nil or diff is not supported, this field will be nil.
	Diff *string

	// BaseEditorConfig and ExportEditorConfig type depends on plugin, should set by plugin

	// If BaseEditorConfig and ExportEditorConfig set, and Diff is not set, export exportService will calculate diff for plugin by using json diff
	// BaseEditorConfig contains the original editor-specific configuration before export.
	BaseEditorConfig any
	// ExportEditorConfig contains the editor-specific configuration after export.
	ExportEditorConfig any
}

// ConfigDetectOptions provides configuration for a config detect operation.
type ConfigDetectOptions struct {
	// Whether to in sandbox mode, in sandbox mode, shell command lookup is not effective, like `code` command for vscode can not be found
	Sandbox bool
}
