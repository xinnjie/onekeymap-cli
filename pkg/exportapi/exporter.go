package exportapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// ExportOptions provides configuration for an export operation.
type ExportOptions struct {
	// Editor type
	EditorType pluginapi.EditorType
	// Optional, existing base keymap for specific editor
	Base io.Reader
	// TODO(xinnjie): export api level enum is not a good idea
	DiffType keymapv1.ExportKeymapRequest_DiffType
	// file path for the keymap config
	FilePath string
}

// Exporter defines the contract for converting a universal KeymapSetting
// into an editor-specific format.
type Exporter interface {
	// Export converts a KeymapSetting and writes it to a destination stream.
	// It returns a report detailing any issues encountered during the conversion.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting *keymapv1.Keymap,
		opts ExportOptions,
	) (*ExportReport, error)
}

// ExportReport details issues encountered during an export operation.
type ExportReport struct {
	// The diff between the base and the exported keymap.
	Diff string
}
