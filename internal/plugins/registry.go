package plugins

import (
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

// Registry holds a collection of all available editor plugins.
type Registry struct {
	plugins map[pluginapi.EditorType]pluginapi.Plugin
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[pluginapi.EditorType]pluginapi.Plugin),
	}
}

// Register adds a new plugin to the registry.
func (r *Registry) Register(plugin pluginapi.Plugin) {
	r.plugins[plugin.EditorType()] = plugin
}

// Get retrieves a plugin by its name.
func (r *Registry) Get(editorType pluginapi.EditorType) (pluginapi.Plugin, bool) {
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
