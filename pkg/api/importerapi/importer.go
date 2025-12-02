package importerapi

import (
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
)

// Importer defines the interface for the import service, which handles the
// conversion of editor-specific keymaps into the universal format.
type Importer interface {
	// Import converts keymaps from a source stream. It returns the converted
	// settings and a report detailing any conflicts or unmapped actions.
	Import(ctx context.Context, opts ImportOptions) (*ImportResult, error)
}

// ImportOptions provides configuration for an import operation.
type ImportOptions struct {
	// Required, editor type
	EditorType pluginapi.EditorType
	// Required, input stream, contains the keymap config for different editors
	InputStream io.Reader
	// Optional, existing onekeymap base setting
	Base keymap.Keymap
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	// The converted keymap setting.
	Setting keymap.Keymap
	// Any issues that arose during the import process.
	Report *validateapi.ValidationReport

	// The changes to the keymap setting.
	Changes *KeymapChanges

	SkipReport pluginapi.ImportSkipReport
}

// KeymapChanges represents the changes to a keymap setting.
type KeymapChanges struct {
	// The keymaps that are added.
	Add []keymap.Action
	// The keymaps that are removed.
	Remove []keymap.Action
	// The keymaps that are updated.
	Update []KeymapDiff
}

type KeymapDiff struct {
	Before keymap.Action
	After  keymap.Action
}
