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
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestExportZedKeymap(t *testing.T) {
	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name     string
		setting  keymap.Keymap
		wantJSON string
		wantErr  bool
	}{
		{
			name: "export copy keymap",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
			name: "exports multiple keybindings for same action",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
							parseKB("ctrl+c"),
						},
					},
				},
			},
			wantJSON: `[
  {
    "context": "Editor",
    "bindings": {
      "cmd-c": "editor::Copy",
      "ctrl-c": "editor::Copy"
    }
  }
]`,
			wantErr: false,
		},
		{
			name: "correctly exports multiple actions",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{
							parseKB("alt+3"),
						},
					},
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
		{
			name: "falls back to child action when parent not supported",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.parentNotSupported",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+shift+h"),
						},
					},
				},
			},
			wantJSON: `[
			{
				"context": "Editor",
				"bindings": {
					"cmd-shift-h": "child::Action"
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
			p := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())
			var buf bytes.Buffer
			exporter, err := p.Exporter()
			require.NoError(t, err)
			_, err = exporter.Export(
				context.Background(),
				&buf,
				tt.setting,
				pluginapi.PluginExportOption{
					ExistingConfig: nil,
					TargetPlatform: platform.PlatformMacOS, // Use macOS for consistent test results across platforms
				},
			)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tt.wantJSON, buf.String(), "The exported JSON should match the expected one.")
			}
		})
	}
}

func TestExportZedKeymap_NonDestructive(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name           string
		setting        keymap.Keymap
		existingConfig string
		wantJSON       string
	}{
		// Basic destructive export tests
		{
			name: "export copy keymap without existing config",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
		},
		{
			name: "empty existing config behaves as destructive export",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
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
		{
			name: "existing config with trailing commas and comments",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
				},
			},
			existingConfig: `[
  // Editor context with user bindings
  {
    "context": "Editor",
    "bindings": {
      "cmd-x": "custom::UserAction", // trailing comma here
    },
  }, // trailing comma after object
  {
    // Workspace context
    "context": "Workspace",
    "bindings": {
      "cmd-shift-p": "custom::WorkspaceAction",
    },
  }, // final trailing comma
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(mappingConfig, slog.New(slog.NewTextHandler(io.Discard, nil)), metrics.NewNoop())
			exporter, err := p.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{
				ExistingConfig: nil,
				TargetPlatform: platform.PlatformMacOS, // Use macOS for consistent test results across platforms
			}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.setting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.wantJSON, buf.String())
		})
	}
}

func TestExportZedKeymap_OrderByBaseContext(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	existingConfig := `[
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
]`

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			{
				Name: "actions.edit.copy",
				Bindings: []keybinding.Keybinding{
					parseKB("meta+c"),
				},
			},
		},
	}

	p := New(mappingConfig, slog.New(slog.NewTextHandler(io.Discard, nil)), metrics.NewNoop())
	exporter, err := p.Exporter()
	require.NoError(t, err)

	var buf bytes.Buffer
	_, err = exporter.Export(context.Background(), &buf, setting, pluginapi.PluginExportOption{
		ExistingConfig: strings.NewReader(existingConfig),
		TargetPlatform: platform.PlatformMacOS, // Use macOS for consistent test results across platforms
	})
	require.NoError(t, err)

	wantJSON := `[
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
	]`
	assert.JSONEq(t, wantJSON, buf.String())
}
