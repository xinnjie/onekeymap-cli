package pluginapi

import (
	"context"
	"io"

	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

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

	SkipReport SkipReport
}

type SkipReport struct {
	SkipActions []SkipAction
}

type SkipAction struct {
	Action string
	// Error description about why this action is skipped
	Error error
}
