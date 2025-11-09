package pluginapi

import (
	"context"
	"io"

	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type PluginImportOption struct {
}

type PluginImporter interface {
	// Import reads the editor's configuration source and converts it into the
	// universal onekeymap KeymapSetting format.
	Import(ctx context.Context, source io.Reader, opts PluginImportOption) (*keymapv1.Keymap, error)
}

// ConfigDetectOptions provides configuration for a config detect operation.
type ConfigDetectOptions struct {
	// Whether to in sandbox mode, in sandbox mode, shell command lookup is not effective, like `code` command for vscode can not be found
	Sandbox bool
}
