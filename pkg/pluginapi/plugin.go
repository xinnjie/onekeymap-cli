package pluginapi

import (
	"context"
	"io"

	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type EditorType string

const (
	EditorTypeVSCode   EditorType = "vscode"
	EditorTypeZed      EditorType = "zed"
	EditorTypeIntelliJ EditorType = "intellij"
	EditorTypeHelix    EditorType = "helix"
)

type PluginImportOption struct {
}

type PluginImporter interface {
	// Import reads the editor's configuration source and converts it into the
	// universal onekeymap KeymapSetting format.
	Import(ctx context.Context, source io.Reader, opts PluginImportOption) (*keymapv1.KeymapSetting, error)
}

type PluginExporter interface {
	// Export takes a universal KeymapSetting and writes it to an editor-specific
	// configuration destination.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting *keymapv1.KeymapSetting,
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

type DefaultConfigPathOptions struct {
	RelativeToHome bool
}

type DefaultConfigPathOption func(*DefaultConfigPathOptions)

func WithRelativeToHome(relative bool) DefaultConfigPathOption {
	return func(o *DefaultConfigPathOptions) {
		o.RelativeToHome = relative
	}
}

// Plugin is the core interface that all editor plugins must implement.
// It defines the contract for importing and exporting keymaps.
type Plugin interface {
	// EditorType returns the unique identifier for the plugin (e.g., "vscode", "zed").
	EditorType() EditorType

	// DefaultConfigPath returns the default path to the editor's configuration file based on the platform.
	// Return multiple paths if the editor has multiple configuration files.
	DefaultConfigPath(opts ...DefaultConfigPathOption) ([]string, error)

	// Importer returns an instance of PluginImporter for the plugin.
	// Return ErrNotSupported if the plugin does not support importing.
	Importer() (PluginImporter, error)

	// Exporter returns an instance of PluginExporter for the plugin.
	// Return ErrNotSupported if the plugin does not support exporting.
	Exporter() (PluginExporter, error)
}
