package exportapi

import (
	"io"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// ExportOptions provides configuration for an export operation.
type ExportOptions struct {
	EditorType pluginapi.EditorType
	Base       io.Reader // Optional base keymap for specific editor
}

// Exporter defines the contract for converting a universal KeymapSetting
// into an editor-specific format.
type Exporter interface {
	// Export converts a KeymapSetting and writes it to a destination stream.
	// It returns a report detailing any issues encountered during the conversion.
	Export(destination io.Writer, setting *keymapv1.KeymapSetting, opts ExportOptions) (*ExportReport, error)
}

// ExportReport details issues encountered during an export operation.
type ExportReport struct {
	// The diff between the base and the exported keymap.
	Diff string
}
