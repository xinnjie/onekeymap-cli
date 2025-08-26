package zed

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestExportZedKeymap(t *testing.T) {
	tests := []struct {
		name     string
		setting  *keymapv1.KeymapSetting
		wantJSON string
		wantErr  bool
	}{
		{
			name: "export copy keymap",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy"
    }
  }
]`,
			wantErr: false,
		},
		{
			name: "correctly exports multiple actions",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
			wantJSON: `[
			{
				"context": "context1",
				"bindings": {
					"alt-3": "command1"
				}
			},
			{
				"context": "context2",
				"bindings": {
					"alt-3": "command2"
				}
			}
			]`,
			wantErr: false,
		},
	}

	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)))
			var buf bytes.Buffer
			exporter, err := p.Exporter()
			require.NoError(t, err)
			_, err = exporter.Export(context.Background(), &buf, tt.setting, pluginapi.PluginExportOption{Base: nil})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.wantJSON, buf.String(), "The exported JSON should match the expected one.")
			}
		})
	}
}

func TestExportZedKeymap_NonDestructive(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	tests := []struct {
		name           string
		setting        *keymapv1.KeymapSetting
		existingConfig string
		wantJSON       string
	}{
		// Basic destructive export tests
		{
			name: "export copy keymap without existing config",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: "",
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy"
    }
  }
]`,
		},
		// Non-destructive export tests
		{
			name: "non-destructive export preserves user keybindings",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-x": "custom::UserAction"
    }
  },
  {
    "context": "Workspace",
    "bindings": {
      "cmd-shift-p": "custom::WorkspaceAction"
    }
  }
]`,
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy",
      "cmd-x": "custom::UserAction"
    }
  },
  {
    "context": "Workspace",
    "bindings": {
      "cmd-shift-p": "custom::WorkspaceAction"
    }
  }
]`,
		},
		{
			name: "managed keybinding takes priority over conflicting user keybinding",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "custom::ConflictingAction",
      "cmd-x": "custom::UserAction"
    }
  }
]`,
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy",
      "cmd-x": "custom::UserAction"
    }
  }
]`,
		},
		{
			name: "multiple contexts with mixed conflicts",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "custom::ConflictingCopy",
      "cmd-x": "custom::UserCut"
    }
  },
  {
    "context": "Workspace",
    "bindings": {
      "cmd-shift-p": "custom::WorkspaceAction"
    }
  },
  {
    "context": "Terminal",
    "bindings": {
      "ctrl-c": "custom::TerminalAction"
    }
  }
]`,
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy",
      "cmd-x": "custom::UserCut"
    }
  },
  {
    "context": "Terminal",
    "bindings": {
      "ctrl-c": "custom::TerminalAction"
    }
  },
  {
    "context": "Workspace",
    "bindings": {
      "cmd-shift-p": "custom::WorkspaceAction"
    }
  }
]`,
		},
		{
			name: "empty existing config behaves as destructive export",
			setting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					keymap.NewBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[]`,
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy"
    }
  }
]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(mappingConfig, slog.New(slog.NewTextHandler(io.Discard, nil)))
			exporter, err := p.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{Base: nil}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.setting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.wantJSON, buf.String())
		})
	}
}
