package internal_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/xinnjie/onekeymap-cli/internal"
	vscodeplugin "github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
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
						Command: "editor.action.formatSelection",
						When:    "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly",
					},
					{
						EditorActionMapping: mappings.EditorActionMapping{DisableImport: true},
						Command:             "notebook.formatCell",
						When:                "editorHasDocumentFormattingProvider && editorTextFocus && inCompositeEditor && notebookEditable && !editorReadonly && activeEditor == 'workbench.editor.notebook'",
					},
				},
			},
		},
	}

	// Use real VSCode plugin importer for this scenario
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := plugins.NewRegistry()
	registry.Register(vscodeplugin.New(mappingConfig, logger, metrics.NewNoop()))
	service := internal.NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

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
	base := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.edit.formatSelection", "ctrl+shift+alt+l"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: bytes.NewReader(vscodeJSON),
		Base:        base,
	}

	res, err := service.Import(t.Context(), opts)
	require.NoError(t, err)
	require.NotNil(t, res)

	expected := &importapi.ImportResult{
		Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
			{
				Name: "actions.edit.formatSelection",
				ActionConfig: &keymapv1.ActionConfig{
					DisplayName: "Format selection",
					Description: "Format Selection",
					Category:    "Editor",
				},
				Bindings: []*keymapv1.KeybindingReadable{
					{
						KeyChords:         keymap.MustParseKeyBinding("ctrl+shift+alt+l").KeyChords,
						KeyChordsReadable: "ctrl+shift+alt+l",
					},
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
