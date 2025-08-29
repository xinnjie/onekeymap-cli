package importapi

import (
	"context"
	"io"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// ImportOptions provides configuration for an import operation.
type ImportOptions struct {
	EditorType  pluginapi.EditorType
	InputStream io.Reader               // Required: input stream, contains the keymap config for different editors
	Base        *keymapv1.KeymapSetting // Optional: base keymap setting
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	// The converted keymap setting.
	Setting *keymapv1.KeymapSetting
	// Any issues that arose during the import process.
	Report *keymapv1.ValidationReport

	// The changes to the keymap setting.
	Changes *KeymapChanges
}

// KeymapChanges represents the changes to a keymap setting.
type KeymapChanges struct {
	// The keymaps that are added.
	Add []*keymapv1.KeyBinding
	// The keymaps that are removed.
	Remove []*keymapv1.KeyBinding
	// The keymaps that are updated.
	Update []KeymapDiff
}

type KeymapDiff struct {
	Before *keymapv1.KeyBinding
	After  *keymapv1.KeyBinding
}

// Importer defines the interface for the import service, which handles the
// conversion of editor-specific keymaps into the universal format.
type Importer interface {
	// Import converts keymaps from a source stream. It returns the converted
	// settings and a report detailing any conflicts or unmapped actions.
	Import(ctx context.Context, opts ImportOptions) (*ImportResult, error)
}
