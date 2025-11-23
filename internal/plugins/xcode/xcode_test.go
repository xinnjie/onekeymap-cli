package xcode_test

import (
	"log/slog"
	"testing"

	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestXcodePlugin_EditorType(t *testing.T) {
	plugin := xcode.New(&mappings.MappingConfig{}, slog.Default(), metrics.NewNoop())

	if got := plugin.EditorType(); got != pluginapi.EditorTypeXcode {
		t.Errorf("EditorType() = %v, want %v", got, pluginapi.EditorTypeXcode)
	}
}

func TestXcodePlugin_Importer(t *testing.T) {
	plugin := xcode.New(&mappings.MappingConfig{}, slog.Default(), metrics.NewNoop())

	importer, err := plugin.Importer()
	if err != nil {
		t.Errorf("Importer() error = %v", err)
		return
	}
	if importer == nil {
		t.Error("Importer() returned nil")
	}
}

func TestXcodePlugin_Exporter(t *testing.T) {
	plugin := xcode.New(&mappings.MappingConfig{}, slog.Default(), metrics.NewNoop())

	exporter, err := plugin.Exporter()
	if err != nil {
		t.Errorf("Exporter() error = %v", err)
		return
	}
	if exporter == nil {
		t.Error("Exporter() returned nil")
	}
}
