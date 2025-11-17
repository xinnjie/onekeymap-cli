package intellij

import (
	"fmt"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type intellijVariantPlugin struct {
	*intellijPlugin

	editorType pluginapi2.EditorType
}

func newIntellijVariantPlugin(
	editorType pluginapi2.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return &intellijVariantPlugin{
		intellijPlugin: newIntellijPlugin(mappingConfig, logger, recorder),
		editorType:     editorType,
	}
}

// EditorType implements pluginapi.Plugin.
func (p *intellijVariantPlugin) EditorType() pluginapi2.EditorType {
	return p.editorType
}

// Importer implements pluginapi.Plugin.
func (p *intellijVariantPlugin) Importer() (pluginapi2.PluginImporter, error) {
	return p.intellijPlugin.Importer()
}

// Exporter implements pluginapi.Plugin.
func (p *intellijVariantPlugin) Exporter() (pluginapi2.PluginExporter, error) {
	return p.intellijPlugin.Exporter()
}

// ConfigDetect implements pluginapi.Plugin.
func (p *intellijVariantPlugin) ConfigDetect(
	opts pluginapi2.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	switch p.editorType {
	case pluginapi2.EditorTypePyCharm:
		return detectConfigForIDE("PyCharm", "PyCharm*", "pycharm", opts)
	case pluginapi2.EditorTypeIntelliJCommunity:
		return detectConfigForIDE("IntelliJ IDEA Community", "IdeaIC*", "idea1", opts)
	case pluginapi2.EditorTypeWebStorm:
		return detectConfigForIDE("WebStorm", "WebStorm*", "webstorm", opts)
	case pluginapi2.EditorTypeClion:
		return detectConfigForIDE("CLion", "CLion*", "clion", opts)
	case pluginapi2.EditorTypePhpStorm:
		return detectConfigForIDE("PhpStorm", "PhpStorm*", "phpstorm", opts)
	case pluginapi2.EditorTypeRubyMine:
		return detectConfigForIDE("RubyMine", "RubyMine*", "rubymine", opts)
	case pluginapi2.EditorTypeGoLand:
		return detectConfigForIDE("GoLand", "GoLand*", "goland", opts)
	case pluginapi2.EditorTypeRustRover:
		return detectConfigForIDE("RustRover", "RustRover*", "rustrover", opts)
	case pluginapi2.EditorTypeIntelliJ:
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
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypePyCharm, mappingConfig, logger, recorder)
}

// NewIntelliJCommunity creates an IntelliJ Community plugin instance.
func NewIntelliJCommunity(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeIntelliJCommunity, mappingConfig, logger, recorder)
}

// NewWebStorm creates a WebStorm plugin instance.
func NewWebStorm(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeWebStorm, mappingConfig, logger, recorder)
}

// NewClion creates a CLion plugin instance.
func NewClion(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeClion, mappingConfig, logger, recorder)
}

// NewPhpStorm creates a PhpStorm plugin instance.
func NewPhpStorm(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypePhpStorm, mappingConfig, logger, recorder)
}

// NewRubyMine creates a RubyMine plugin instance.
func NewRubyMine(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeRubyMine, mappingConfig, logger, recorder)
}

// NewGoLand creates a GoLand plugin instance.
func NewGoLand(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeGoLand, mappingConfig, logger, recorder)
}

// NewRustRover creates a RustRover plugin instance.
func NewRustRover(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newIntellijVariantPlugin(pluginapi2.EditorTypeRustRover, mappingConfig, logger, recorder)
}
