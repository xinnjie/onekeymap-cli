package importapi

import (
	"context"
	"io"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// ImportOptions provides configuration for an import operation.
type ImportOptions struct {
	// Required, editor type
	EditorType pluginapi.EditorType
	// Required, input stream, contains the keymap config for different editors
	InputStream io.Reader
	// Optional, existing onekeymap base setting
	Base *keymapv1.Keymap
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	// The converted keymap setting.
	Setting *keymapv1.Keymap
	// Any issues that arose during the import process.
	Report *keymapv1.ValidationReport

	// The changes to the keymap setting.
	Changes *KeymapChanges
}

// KeymapChanges represents the changes to a keymap setting.
type KeymapChanges struct {
	// The keymaps that are added.
	Add []*keymapv1.ActionBinding
	// The keymaps that are removed.
	Remove []*keymapv1.ActionBinding
	// The keymaps that are updated.
	Update []KeymapDiff
}

type KeymapDiff struct {
	Before *keymapv1.ActionBinding
	After  *keymapv1.ActionBinding
}

// Importer defines the interface for the import service, which handles the
// conversion of editor-specific keymaps into the universal format.
type Importer interface {
	// Import converts keymaps from a source stream. It returns the converted
	// settings and a report detailing any conflicts or unmapped actions.
	Import(ctx context.Context, opts ImportOptions) (*ImportResult, error)
}
