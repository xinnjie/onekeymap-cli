package exporterapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// Exporter defines the contract for converting a universal KeymapSetting
// into an editor-specific format.
type Exporter interface {
	// Export converts a KeymapSetting and writes it to a destination stream.
	// It returns a report detailing any issues encountered during the conversion.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting keymap.Keymap,
		opts ExportOptions,
	) (*ExportReport, error)
}

// DiffType represents the type of diff output format.
type DiffType int

const (
	// DiffTypeUnspecified represents an unspecified diff type.
	DiffTypeUnspecified DiffType = iota
	// DiffTypeASCII represents ASCII diff format.
	DiffTypeASCII
	// DiffTypeUnified represents unified diff format.
	DiffTypeUnified
)

// String returns the string representation of the DiffType.
func (d DiffType) String() string {
	switch d {
	case DiffTypeUnspecified:
		return "UNSPECIFIED"
	case DiffTypeASCII:
		return "ASCII"
	case DiffTypeUnified:
		return "UNIFIED"
	default:
		return "UNKNOWN"
	}
}

// ExportOptions provides configuration for an export operation.
type ExportOptions struct {
	// Editor type
	EditorType pluginapi.EditorType
	// Optional, existing original keymap for specific editor before export
	OriginalConfig io.Reader
	// Use diff type to generate ascii diff text or unified diff text
	DiffType DiffType
	// file path for the keymap config
	FilePath string
}

// ExportReport details issues encountered during an export operation.
type ExportReport struct {
	// The diff text between the original keymap and keymap after export.
	Diff string

	// Coverage reports how many actions were successfully exported.
	Coverage ExportCoverage

	// SkipActions reports actions that were not exported and why.
	SkipActions []pluginapi.ExportSkipAction
}

// ExportCoverage summarizes export success rate.
type ExportCoverage struct {
	// TotalActions is the number of actions requested for export.
	TotalActions int
	// FullyExported is the number of actions where all keybindings were exported.
	FullyExported int
	// PartiallyExported lists actions where some keybindings could not be exported.
	PartiallyExported []PartialExportedAction
}

// PartialExportedAction describes an action that was only partially exported.
type PartialExportedAction struct {
	// Action name, e.g. "actions.clipboard.copy"
	Action string
	// Requested keybindings that were requested to be exported
	Requested []keybinding.Keybinding
	// Exported keybindings that were actually exported
	Exported []keybinding.Keybinding
	// Reason explains why some keybindings were not exported
	Reason string
}
