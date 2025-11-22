package registry

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/basekeymap"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/helix"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/zed"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
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

func NewRegistryWithPlugins(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *Registry {
	r := NewRegistry()

	// VSCode family
	r.Register(vscode.New(mappingConfig, logger, recorder))
	r.Register(vscode.NewWindsurf(mappingConfig, logger, recorder))
	r.Register(vscode.NewWindsurfNext(mappingConfig, logger, recorder))
	r.Register(vscode.NewCursor(mappingConfig, logger, recorder))

	// IntelliJ family
	r.Register(intellij.New(mappingConfig, logger, recorder))
	r.Register(intellij.NewPycharm(mappingConfig, logger, recorder))
	r.Register(intellij.NewIntelliJCommunity(mappingConfig, logger, recorder))
	r.Register(intellij.NewWebStorm(mappingConfig, logger, recorder))
	r.Register(intellij.NewClion(mappingConfig, logger, recorder))
	r.Register(intellij.NewPhpStorm(mappingConfig, logger, recorder))
	r.Register(intellij.NewRubyMine(mappingConfig, logger, recorder))
	r.Register(intellij.NewGoLand(mappingConfig, logger, recorder))
	r.Register(intellij.NewRustRover(mappingConfig, logger, recorder))

	r.Register(helix.New(mappingConfig, logger))
	r.Register(zed.New(mappingConfig, logger, recorder))
	r.Register(xcode.New(mappingConfig, logger, recorder))

	r.Register(basekeymap.New())
	return r
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
