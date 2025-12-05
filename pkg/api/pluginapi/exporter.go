package pluginapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	keybinding "github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

type PluginExporter interface {
	// Export takes a universal KeymapSetting and writes it to an editor-specific
	// configuration destination.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting keymap.Keymap,
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

	// TargetPlatform specifies the target platform for the exported keybindings.
	// This affects how modifier keys are formatted (e.g., "cmd" for macOS, "win" for Windows).
	// If empty, defaults to the current runtime platform.
	TargetPlatform platform.Platform
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

	// SkipReport reports actions that were not exported and why.
	SkipReport ExportSkipReport

	ExportedReport ExportedReport
}

type ExportSkipReport struct {
	SkipActions []ExportSkipAction
}

type ExportSkipAction struct {
	// Action name, e.g. "actions.clipboard.copy"
	Action string
	// Error description about why this action is skipped
	Error error
}

// ExportedReport contains detailed export results for each action.
// This is used by the exporter service to compute ExportCoverage.
type ExportedReport struct {
	Actions []ActionExportResult
}

// ActionExportResult describes the export outcome for a single action.
type ActionExportResult struct {
	// Action name, e.g. "actions.clipboard.copy"
	Action string
	// Requested keybindings that were requested to be exported
	Requested []keybinding.Keybinding
	// Exported keybindings that were actually exported (may be fewer due to editor limitations)
	Exported []keybinding.Keybinding
	// Reason explains why some keybindings were not exported (optional)
	Reason string
}
