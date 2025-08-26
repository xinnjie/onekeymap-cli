package zed

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func testSettingCopy() *keymapv1.KeymapSetting {
	return &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			{
				Action: "actions.edit.copy",
				KeyChords: &keymapv1.KeyChordSequence{
					Chords: []*keymapv1.KeyChord{
						{
							KeyCode:   "c",
							Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_META},
						},
					},
				},
			},
		},
	}
}

const expectedCopyJSON = `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy"
    }
  }
]`

const baseCopyJSONChanged = `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Paste"
    }
  }
]`

func TestZedExporter_Diff_NilBase_ShouldHaveAdditions(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	exporter := newExporter(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), diff.NewJsonDiffer())

	var buf bytes.Buffer
	report, err := exporter.Export(context.Background(), &buf, testSettingCopy(), pluginapi.PluginExportOption{Base: nil})
	require.NoError(t, err)
	assert.JSONEq(t, expectedCopyJSON, buf.String())
	assert.NotEmpty(t, *report.Diff, "diff should show additions when base is empty")
	// minimal shape checks for added array element at index 0 and presence of the action string
	assert.Contains(t, *report.Diff, "\n+  0:")
	assert.Contains(t, *report.Diff, "editor::Copy")
}

func TestZedExporter_Diff_SameBase_ShouldBeEmpty(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	exporter := newExporter(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), diff.NewJsonDiffer())

	base := bytes.NewBufferString(expectedCopyJSON)
	var buf bytes.Buffer
	report, err := exporter.Export(context.Background(), &buf, testSettingCopy(), pluginapi.PluginExportOption{Base: base})
	require.NoError(t, err)
	assert.JSONEq(t, expectedCopyJSON, buf.String())
	assert.Equal(t, "", *report.Diff, "diff should be empty when base equals output")
}

func TestZedExporter_Diff_NonEmptyBase_ShouldShowModifications(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	exporter := newExporter(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), diff.NewJsonDiffer())

	base := bytes.NewBufferString(baseCopyJSONChanged)
	var buf bytes.Buffer
	report, err := exporter.Export(context.Background(), &buf, testSettingCopy(), pluginapi.PluginExportOption{Base: base})
	require.NoError(t, err)
	// Output should still be expectedCopyJSON
	assert.JSONEq(t, expectedCopyJSON, buf.String())
	// Diff should contain a removal of Paste and an addition of Copy for cmd-c
	require.NotEmpty(t, *report.Diff)
	t.Logf("diff output:\n%s", *report.Diff)
	assert.Contains(t, *report.Diff, "\n-", "should indicate removal in diff")
	assert.Contains(t, *report.Diff, "\n+", "should indicate addition in diff")
	assert.Contains(t, *report.Diff, "editor::Paste", "diff should reference old action")
	assert.Contains(t, *report.Diff, "editor::Copy", "diff should reference new action")
}

func TestZedExporter_Diff_InvalidBase_ShouldError(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	exporter := newExporter(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), diff.NewJsonDiffer())

	base := bytes.NewBufferString("{ invalid json ")
	var buf bytes.Buffer
	_, err = exporter.Export(context.Background(), &buf, testSettingCopy(), pluginapi.PluginExportOption{Base: base})
	assert.Error(t, err)
}
