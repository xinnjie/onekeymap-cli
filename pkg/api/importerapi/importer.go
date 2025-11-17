package importerapi

import (
	"context"
	"io"

	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// ImportOptions provides configuration for an import operation.
type ImportOptions struct {
	// Required, editor type
	EditorType pluginapi2.EditorType
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

	SkipReport pluginapi2.ImportSkipReport
}

// KeymapChanges represents the changes to a keymap setting.
type KeymapChanges struct {
	// The keymaps that are added.
	Add []*keymapv1.Action
	// The keymaps that are removed.
	Remove []*keymapv1.Action
	// The keymaps that are updated.
	Update []KeymapDiff
}

type KeymapDiff struct {
	Before *keymapv1.Action
	After  *keymapv1.Action
}

// Importer defines the interface for the import service, which handles the
// conversion of editor-specific keymaps into the universal format.
type Importer interface {
	// Import converts keymaps from a source stream. It returns the converted
	// settings and a report detailing any conflicts or unmapped actions.
	Import(ctx context.Context, opts ImportOptions) (*ImportResult, error)
}
