package basekeymap

import (
	"bytes"
	"context"
	"io"

	"github.com/xinnjie/onekeymap-cli/config/base"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type plugin struct{}

func New() pluginapi.Plugin {
	return &plugin{}
}

func (p *plugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeBasekeymap
}

func (p *plugin) ConfigDetect(_ pluginapi.ConfigDetectOptions) ([]string, bool, error) {
	bases, err := base.List()
	if err != nil {
		return nil, false, err
	}
	return bases, true, nil
}

func (p *plugin) Importer() (pluginapi.PluginImporter, error) {
	return &importer{}, nil
}

func (p *plugin) Exporter() (pluginapi.PluginExporter, error) {
	return nil, pluginapi.ErrNotSupported
}

type importer struct{}

func (i *importer) Import(
	_ context.Context,
	source io.Reader,
	_ pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	// Read the base keymap name from source
	nameBytes, err := io.ReadAll(source)
	if err != nil {
		return pluginapi.PluginImportResult{}, err
	}
	name := string(bytes.TrimSpace(nameBytes))

	// Read the embedded base keymap JSON
	data, err := base.Read(name)
	if err != nil {
		return pluginapi.PluginImportResult{}, err
	}

	// Parse and return the keymap
	km, err := keymap.Load(bytes.NewReader(data), keymap.LoadOptions{})
	if err != nil {
		return pluginapi.PluginImportResult{}, err
	}

	return pluginapi.PluginImportResult{Keymap: km}, nil
}
