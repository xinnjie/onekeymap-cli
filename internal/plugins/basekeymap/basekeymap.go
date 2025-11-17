package basekeymap

import (
	"bytes"
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/config/base"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type plugin struct{}

func New() pluginapi2.Plugin {
	return &plugin{}
}

func (p *plugin) EditorType() pluginapi2.EditorType {
	return pluginapi2.EditorTypeBasekeymap
}

func (p *plugin) ConfigDetect(_ pluginapi2.ConfigDetectOptions) ([]string, bool, error) {
	bases, err := base.List()
	if err != nil {
		return nil, false, err
	}
	return bases, true, nil
}

func (p *plugin) Importer() (pluginapi2.PluginImporter, error) {
	return &importer{}, nil
}

func (p *plugin) Exporter() (pluginapi2.PluginExporter, error) {
	return nil, pluginapi2.ErrNotSupported
}

type importer struct{}

func (i *importer) Import(
	_ context.Context,
	source io.Reader,
	_ pluginapi2.PluginImportOption,
) (pluginapi2.PluginImportResult, error) {
	// Read the base keymap name from source
	nameBytes, err := io.ReadAll(source)
	if err != nil {
		return pluginapi2.PluginImportResult{}, err
	}
	name := string(bytes.TrimSpace(nameBytes))

	// Read the embedded base keymap JSON
	data, err := base.Read(name)
	if err != nil {
		return pluginapi2.PluginImportResult{}, err
	}

	// Parse and return the keymap
	km, err := keymap.Load(bytes.NewReader(data), keymap.LoadOptions{})
	if err != nil {
		return pluginapi2.PluginImportResult{}, err
	}

	return pluginapi2.PluginImportResult{Keymap: km}, nil
}
