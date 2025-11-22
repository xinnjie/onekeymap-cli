package importer_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	vscodeplugin "github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/importer"
	mappings2 "github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

func TestImportEndToEnd_Import_VSCode_FormatSelection_NoChange(t *testing.T) {
	// Setup mapping config according to provided YAML
	mappingConfig := &mappings2.MappingConfig{
		Mappings: map[string]mappings2.ActionMappingConfig{
			"actions.edit.formatSelection": {
				ID:          "actions.edit.formatSelection",
				Name:        "Format selection",
				Description: "Format Selection",
				Category:    "Editor",
				VSCode: mappings2.VscodeConfigs{
					{
						Command: "editor.action.formatSelection",
						When:    "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly",
					},
					{
						EditorActionMapping: mappings2.EditorActionMapping{DisableImport: true},
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
	service := importer.NewImporter(registry, mappingConfig, logger, metrics.NewNoop())

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
	baseBinding, _ := keybinding.NewKeybinding("ctrl+shift+alt+l", keybinding.ParseOption{Separator: "+"})
	base := keymap.Keymap{
		Actions: []keymap.Action{
			{Name: "actions.edit.formatSelection", Bindings: []keybinding.Keybinding{baseBinding}},
		},
	}

	opts := importerapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: bytes.NewReader(vscodeJSON),
		Base:        base,
	}

	res, err := service.Import(t.Context(), opts)
	require.NoError(t, err)
	require.NotNil(t, res)

	expectedBinding, _ := keybinding.NewKeybinding("ctrl+shift+alt+l", keybinding.ParseOption{Separator: "+"})
	expected := &importerapi.ImportResult{
		Setting: keymap.Keymap{Actions: []keymap.Action{
			{Name: "actions.edit.formatSelection", Bindings: []keybinding.Keybinding{expectedBinding}},
		}},
		Changes: &importerapi.KeymapChanges{},
	}

	settingDiff := cmp.Diff(expected.Setting, res.Setting)
	assert.Empty(t, settingDiff, "Setting mismatch: %s", settingDiff)
	changesDiff := cmp.Diff(expected.Changes, res.Changes)
	assert.Empty(t, changesDiff, "Changes mismatch: %s", changesDiff)
}
