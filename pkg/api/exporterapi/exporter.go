package exporterapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

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
	// Optional, existing base keymap for specific editor
	Base io.Reader
	// TODO(xinnjie): export api level enum is not a good idea
	DiffType DiffType
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
		setting keymap.Keymap,
		opts ExportOptions,
	) (*ExportReport, error)
}

// ExportReport details issues encountered during an export operation.
type ExportReport struct {
	// The diff between the base and the exported keymap.
	Diff string

	// SkipActions reports actions that were not exported and why.
	SkipActions []pluginapi.ExportSkipAction
}
