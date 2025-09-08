package onekeymap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	vscodeplugin "github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/vscode"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestImportEndToEnd_Import_VSCode_FormatSelection_NoChange(t *testing.T) {
	// Setup mapping config according to provided YAML
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.edit.formatSelection": {
				ID:          "actions.edit.formatSelection",
				Name:        "Format selection",
				Description: "Format Selection",
				Category:    "Editor",
				VSCode: mappings.VscodeConfigs{
					{
						EditorActionMapping: mappings.EditorActionMapping{ForImport: true},
						Command:             "editor.action.formatSelection",
						When:                "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly",
					},
					{
						Command: "notebook.formatCell",
						When:    "editorHasDocumentFormattingProvider && editorTextFocus && inCompositeEditor && notebookEditable && !editorReadonly && activeEditor == 'workbench.editor.notebook'",
					},
				},
			},
		},
	}

	// Use real VSCode plugin importer for this scenario
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := plugins.NewRegistry()
	registry.Register(vscodeplugin.New(mappingConfig, logger))
	service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

	// VSCode keybindings.json content (comments stripped by importer)
	// Note that "ctrl+alt+shift+l" is different from "ctrl+shift+alt+l"
	vscodeJSON := []byte(`[
  {
    "key": "ctrl+alt+shift+l",
    "command": "editor.action.formatSelection",
    "when": "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly"
  }
]`)

	// Base config has the same binding (order of modifiers irrelevant; parser normalizes)
	base := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.formatSelection", "ctrl+shift+alt+l"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: bytes.NewReader(vscodeJSON),
		Base:        base,
	}

	res, err := service.Import(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, res)

	expected := &importapi.ImportResult{
		Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
			{
				Id:          "actions.edit.formatSelection",
				Name:        "Format selection",
				Description: "Format Selection",
				Category:    "Editor",
				Bindings: []*keymapv1.Binding{
					{KeyChords: keymap.MustParseKeyBinding("ctrl+shift+alt+l").KeyChords, KeyChordsReadable: "ctrl+shift+alt+l"},
				},
			},
		}},
		Changes: &importapi.KeymapChanges{},
	}

	settingDiff := cmp.Diff(expected.Setting, res.Setting, protocmp.Transform())
	assert.Empty(t, settingDiff)
	changesDiff := cmp.Diff(expected.Changes, res.Changes, protocmp.Transform())
	assert.Empty(t, changesDiff)
}
