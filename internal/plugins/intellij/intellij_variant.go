package intellij

import (
	"fmt"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

type intellijVariantPlugin struct {
	*intellijPlugin

	editorType pluginapi.EditorType
}

func newIntellijVariantPlugin(
	editorType pluginapi.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return &intellijVariantPlugin{
		intellijPlugin: newIntellijPlugin(mappingConfig, logger, recorder),
		editorType:     editorType,
	}
}

// EditorType implements pluginapi.Plugin.
func (p *intellijVariantPlugin) EditorType() pluginapi.EditorType {
	return p.editorType
}

// Importer implements pluginapi.Plugin.
func (p *intellijVariantPlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.intellijPlugin.Importer()
}

// Exporter implements pluginapi.Plugin.
func (p *intellijVariantPlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.intellijPlugin.Exporter()
}

// ConfigDetect implements pluginapi.Plugin.
func (p *intellijVariantPlugin) ConfigDetect(
	opts pluginapi.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	switch p.editorType {
	case pluginapi.EditorTypePyCharm:
		return detectConfigForIDE("PyCharm", "PyCharm*", "pycharm", opts)
	case pluginapi.EditorTypeIntelliJCommunity:
		return detectConfigForIDE("IntelliJ IDEA Community", "IdeaIC*", "idea1", opts)
	case pluginapi.EditorTypeWebStorm:
		return detectConfigForIDE("WebStorm", "WebStorm*", "webstorm", opts)
	case pluginapi.EditorTypeClion:
		return detectConfigForIDE("CLion", "CLion*", "clion", opts)
	case pluginapi.EditorTypePhpStorm:
		return detectConfigForIDE("PhpStorm", "PhpStorm*", "phpstorm", opts)
	case pluginapi.EditorTypeRubyMine:
		return detectConfigForIDE("RubyMine", "RubyMine*", "rubymine", opts)
	case pluginapi.EditorTypeGoLand:
		return detectConfigForIDE("GoLand", "GoLand*", "goland", opts)
	case pluginapi.EditorTypeRustRover:
		return detectConfigForIDE("RustRover", "RustRover*", "rustrover", opts)
	case pluginapi.EditorTypeIntelliJ:
		return detectConfigForIDE("IntelliJ IDEA", "IntelliJIdea*", "idea", opts)
	default:
		return nil, false, fmt.Errorf("unknown editor type: %s", p.editorType)
	}
}

// NewPycharm creates a PyCharm plugin instance.
func NewPycharm(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypePyCharm, mappingConfig, logger, recorder)
}

// NewIntelliJCommunity creates an IntelliJ Community plugin instance.
func NewIntelliJCommunity(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeIntelliJCommunity, mappingConfig, logger, recorder)
}

// NewWebStorm creates a WebStorm plugin instance.
func NewWebStorm(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeWebStorm, mappingConfig, logger, recorder)
}

// NewClion creates a CLion plugin instance.
func NewClion(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeClion, mappingConfig, logger, recorder)
}

// NewPhpStorm creates a PhpStorm plugin instance.
func NewPhpStorm(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypePhpStorm, mappingConfig, logger, recorder)
}

// NewRubyMine creates a RubyMine plugin instance.
func NewRubyMine(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeRubyMine, mappingConfig, logger, recorder)
}

// NewGoLand creates a GoLand plugin instance.
func NewGoLand(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeGoLand, mappingConfig, logger, recorder)
}

// NewRustRover creates a RustRover plugin instance.
func NewRustRover(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newIntellijVariantPlugin(pluginapi.EditorTypeRustRover, mappingConfig, logger, recorder)
}
