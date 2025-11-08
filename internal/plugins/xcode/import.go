package xcode

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"howett.net/plist"
)

// xcodeImporter handles importing keybindings from Xcode.
type xcodeImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	reporter      *metrics.UnknownActionReporter
}

func newImporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *xcodeImporter {
	return &xcodeImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		reporter:      metrics.NewUnknownActionReporter(recorder),
	}
}

// Import reads an Xcode .idekeybindings file and converts it to a universal KeymapSetting.
func (i *xcodeImporter) Import(
	ctx context.Context,
	source io.Reader,
	_ pluginapi.PluginImportOption,
) (*keymapv1.Keymap, error) {
	// Read the plist XML data
	xmlData, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	plistData, err := parseXcodeConfig(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xcode keybindings: %w", err)
	}

	setting := &keymapv1.Keymap{}
	for _, binding := range plistData.MenuKeyBindings.KeyBindings {
		// Skip bindings without keyboard shortcuts
		if binding.KeyboardShortcut == "" {
			continue
		}

		mapping := i.FindByXcodeAction(binding.Action)
		if mapping == nil {
			i.logger.DebugContext(
				ctx,
				"Skipping keybinding with unknown action",
				"action",
				binding.Action,
				"commandID",
				binding.CommandID,
			)
			i.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeXcode, binding.Action)
			continue
		}

		kb, err := ParseKeybinding(binding.KeyboardShortcut)
		if err != nil {
			i.logger.WarnContext(
				ctx,
				"Skipping keybinding with unparsable key",
				"key",
				binding.KeyboardShortcut,
				"error",
				err,
			)
			continue
		}

		newKeymap := &keymapv1.Action{
			Name:     mapping.ID,
			Bindings: []*keymapv1.KeybindingReadable{{KeyChords: kb.KeyChords}},
		}
		setting.Actions = append(setting.Actions, newKeymap)
	}

	// Process Text Key Bindings
	for key, val := range plistData.TextKeyBindings.KeyBindings {
		// Only import when there is exactly one text action; skip arrays and empty
		if len(val.Items) != 1 {
			continue
		}
		textAction := val.Items[0]

		mapping := i.FindByXcodeTextAction(textAction)
		if mapping == nil {
			i.logger.DebugContext(
				ctx,
				"Skipping text keybinding with unknown action",
				"textAction",
				textAction,
			)
			i.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeXcode, textAction)
			continue
		}

		kb, err := ParseKeybinding(key)
		if err != nil {
			i.logger.WarnContext(
				ctx,
				"Skipping text keybinding with unparsable key",
				"key",
				key,
				"error",
				err,
			)
			continue
		}

		newKeymap := &keymapv1.Action{
			Name:     mapping.ID,
			Bindings: []*keymapv1.KeybindingReadable{{KeyChords: kb.KeyChords}},
		}
		setting.Actions = append(setting.Actions, newKeymap)
	}

	setting.Actions = internal.DedupKeyBindings(setting.GetActions())

	return setting, nil
}

// parseXcodeConfig parses the plist format and extracts keybindings using go-plist
func parseXcodeConfig(xmlData []byte) (*xcodeKeybindingsPlist, error) {
	var plistData xcodeKeybindingsPlist

	// Use plist library to decode the plist data
	_, err := plist.Unmarshal(xmlData, &plistData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode plist: %w", err)
	}

	return &plistData, nil
}
