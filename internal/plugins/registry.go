package plugins

import (
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// Registry holds a collection of all available editor plugins.
type Registry struct {
	plugins map[pluginapi2.EditorType]pluginapi2.Plugin
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[pluginapi2.EditorType]pluginapi2.Plugin),
	}
}

// Register adds a new plugin to the registry.
func (r *Registry) Register(plugin pluginapi2.Plugin) {
	r.plugins[plugin.EditorType()] = plugin
}

// Get retrieves a plugin by its name.
func (r *Registry) Get(editorType pluginapi2.EditorType) (pluginapi2.Plugin, bool) {
	plugin, ok := r.plugins[editorType]
	return plugin, ok
}

func (r *Registry) GetNames() []string {
	var names []string
	for name := range r.plugins {
		names = append(names, string(name))
	}
	return names
}
