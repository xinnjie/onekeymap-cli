package intellij

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type intellijImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
}

func newImporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger) *intellijImporter {
	return &intellijImporter{mappingConfig: mappingConfig, logger: logger}
}

// Import converts IntelliJ keymap XML into the universal KeymapSetting.
func (p *intellijImporter) Import(ctx context.Context, source io.Reader, opts pluginapi.PluginImportOption) (*keymapv1.KeymapSetting, error) {
	_ = ctx
	_ = opts

	raw, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	var doc KeymapXML
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse intellij keymap xml: %w", err)
	}

	setting := &keymapv1.KeymapSetting{}

	for _, act := range doc.Actions {
		actionID, err := p.ActionIDFromIntelliJ(act.ID)
		if err != nil {
			// Not found in mapping, skip quietly
			p.logger.Debug("no universal mapping for intellij action", "action", act.ID)
			continue
		}

		for _, ks := range act.KeyboardShortcuts {
			if ks.First == "" {
				continue
			}
			kb, err := parseKeyBinding(ks)
			if err != nil {
				p.logger.Warn("failed to parse key binding", "binding", ks, "error", err)
				continue
			}
			newBinding := &keymapv1.KeyBinding{
				Action:    actionID,
				KeyChords: kb.KeyChords,
			}
			setting.Keybindings = append(setting.Keybindings, newBinding)
		}
	}
	return setting, nil
}

// ActionIDFromIntelliJ converts an IntelliJ action and optional context to a universal action ID.
func (p *intellijImporter) ActionIDFromIntelliJ(action string) (string, error) {
	for _, mapping := range p.mappingConfig.Mappings {
		if mapping.IntelliJ.Action == action {
			return mapping.ID, nil
		}
	}
	return "", fmt.Errorf("no mapping found for intellij action: %s", action)
}
