package pluginapi

// Plugin is the core interface that all editor plugins must implement.
// It defines the contract for importing and exporting keymaps.
type Plugin interface {
	// EditorType returns the unique identifier for the plugin (e.g., "vscode", "zed").
	EditorType() EditorType

	// ConfigDetect returns the default path to the editor's configuration file based on the platform.
	// Return multiple paths if the editor has multiple configuration files.
	ConfigDetect(opts ConfigDetectOptions) (paths []string, installed bool, err error)

	// Importer returns an instance of PluginImporter for the plugin.
	// Return ErrNotSupported if the plugin does not support importing.
	Importer() (PluginImporter, error)

	// Exporter returns an instance of PluginExporter for the plugin.
	// Return ErrNotSupported if the plugin does not support exporting.
	Exporter() (PluginExporter, error)
}
