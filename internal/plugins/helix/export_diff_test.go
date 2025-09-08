package helix

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func testSettingCopy() *keymapv1.KeymapSetting {
	return &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
		},
	}
}

func expectedCopyTOMLMap() map[string]any {
	return map[string]any{
		"keys": map[string]any{
			"insert": map[string]any{
				"M-c": "yank",
			},
		},
	}
}

func decodeTOMLMap(t *testing.T, s string) map[string]any {
	var got map[string]any
	require.NoError(t, toml.NewDecoder(bytes.NewBufferString(s)).Decode(&got))
	return got
}

func TestHelixExporter_Diff(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	exporter := newExporter(mappingConfig, slogDiscard(), diff.NewJsonDiffer())

	type baseKind int
	const (
		baseNil baseKind = iota
		baseSame
		baseModified
		baseInvalid
	)

	tests := []struct {
		name             string
		kind             baseKind
		wantErr          bool
		wantDiffEmpty    bool
		wantDiffContains []string
	}{
		{name: "NilBase_ShouldHaveAdditions", kind: baseNil, wantDiffContains: []string{"M-c", "yank", "+"}},
		{name: "SameBase_ShouldBeEmpty", kind: baseSame, wantDiffEmpty: true},
		{name: "NonEmptyBase_ShouldShowModifications", kind: baseModified, wantDiffContains: []string{"-", "+", "paste", "yank"}},
		{name: "InvalidBase_ShouldError", kind: baseInvalid, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var base io.Reader
			switch tc.kind {
			case baseNil:
				base = nil
			case baseSame:
				bb := new(bytes.Buffer)
				require.NoError(t, toml.NewEncoder(bb).Encode(expectedCopyTOMLMap()))
				base = bb
			case baseModified:
				baseMap := map[string]any{
					"keys": map[string]any{
						"insert": map[string]any{
							"M-c": "paste",
						},
					},
				}
				bb := new(bytes.Buffer)
				require.NoError(t, toml.NewEncoder(bb).Encode(baseMap))
				base = bb
			case baseInvalid:
				base = bytes.NewBufferString("invalid toml content")
			}

			var buf bytes.Buffer
			report, err := exporter.Export(context.Background(), &buf, testSettingCopy(), pluginapi.PluginExportOption{Base: base})

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedCopyTOMLMap(), decodeTOMLMap(t, buf.String()))
			if tc.wantDiffEmpty {
				assert.Equal(t, "", *report.Diff, "diff should be empty when base equals output")
			} else {
				assert.NotEmpty(t, *report.Diff)
				for _, s := range tc.wantDiffContains {
					assert.Contains(t, *report.Diff, s)
				}
			}
		})
	}
}

// slogDiscard returns a logger that discards output to keep tests quiet.
// We depend on this in tests so they don't write logs to stdout.
func slogDiscard() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
