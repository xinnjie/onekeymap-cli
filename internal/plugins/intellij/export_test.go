package intellij

import (
	"bytes"
	"context"
	"encoding/xml"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

func TestExportIntelliJKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformLinux, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name           string
		setting        keymap.Keymap
		existingConfig string
		validateFunc   func(t *testing.T, out KeymapXML)
	}{
		// Basic destructive export tests
		{
			name: "basic single-chord export ($Copy)",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
				},
			},
			validateFunc: func(t *testing.T, out KeymapXML) {
				assert.Equal(t, "Onekeymap", out.Name)
				assert.Equal(t, "1", out.Version)
				assert.True(t, out.DisableMnemonics)
				assert.Equal(t, "$default", out.Parent)
				// find $Copy action
				var found *ActionXML
				for i := range out.Actions {
					if out.Actions[i].ID == "$Copy" {
						found = &out.Actions[i]
						break
					}
				}
				if assert.NotNil(t, found, "expected $Copy action in exported XML") {
					if assert.Len(t, found.KeyboardShortcuts, 1) {
						ks := found.KeyboardShortcuts[0]
						assert.Equal(t, "meta C", ks.First)
						assert.Empty(t, ks.Second)
					}
				}
			},
		},
		{
			name: "multi entries and multi-chord export (command1)",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					// single chord ctrl+alt+s
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s")},
					},
					// two-chord ctrl+k ctrl+c
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+k ctrl+c")},
					},
				},
			},
			validateFunc: func(t *testing.T, out KeymapXML) {
				var cmd1 *ActionXML
				for i := range out.Actions {
					if out.Actions[i].ID == "command1" {
						cmd1 = &out.Actions[i]
						break
					}
				}
				if assert.NotNil(t, cmd1, "expected command1 action in exported XML") {
					assert.Len(t, cmd1.KeyboardShortcuts, 2)
					// Order preserved based on input order
					assert.Equal(t, "control alt S", cmd1.KeyboardShortcuts[0].First)
					assert.Empty(t, cmd1.KeyboardShortcuts[0].Second)
					assert.Equal(t, "control K", cmd1.KeyboardShortcuts[1].First)
					assert.Equal(t, "control C", cmd1.KeyboardShortcuts[1].Second)
				}
			},
		},
		{
			name: "unmapped actions skipped",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.unknown",
						Bindings: []keybinding.Keybinding{parseKB("meta+x")},
					},
				},
			},
			validateFunc: func(t *testing.T, out KeymapXML) {
				// No actions should be exported
				assert.Empty(t, out.Actions)
			},
		},
		{
			name: "dedup identical shortcuts for an action",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s")},
					},
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s")},
					},
				},
			},
			validateFunc: func(t *testing.T, out KeymapXML) {
				var cmd1 *ActionXML
				for i := range out.Actions {
					if out.Actions[i].ID == "command1" {
						cmd1 = &out.Actions[i]
						break
					}
				}
				if assert.NotNil(t, cmd1, "expected command1 action in exported XML") {
					// Should only have one shortcut after dedup
					assert.Len(t, cmd1.KeyboardShortcuts, 1)
					assert.Equal(t, "control alt S", cmd1.KeyboardShortcuts[0].First)
					assert.Empty(t, cmd1.KeyboardShortcuts[0].Second)
				}
			},
		},
		{
			name: "special keys are formatted correctly",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("f5")},
					},
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+numpad3")},
					},
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+shift+[")},
					},
				},
			},
			validateFunc: func(t *testing.T, out KeymapXML) {
				var cmd1 *ActionXML
				for i := range out.Actions {
					if out.Actions[i].ID == "command1" {
						cmd1 = &out.Actions[i]
						break
					}
				}
				if assert.NotNil(t, cmd1, "expected command1 action in exported XML") {
					assert.Len(t, cmd1.KeyboardShortcuts, 3)
					// Entries may appear in the order provided
					assert.Equal(t, "F5", cmd1.KeyboardShortcuts[0].First)
					assert.Equal(t, "control NUMPAD3", cmd1.KeyboardShortcuts[1].First)
					assert.Equal(t, "control shift OPEN_BRACKET", cmd1.KeyboardShortcuts[2].First)
				}
			},
		},
		// Non-destructive export tests
		{
			name: "non-destructive export preserves user actions",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<keymap version="1" name="Onekeymap" parent="$default" disable-mnemonics="true">
  <action id="CustomUserAction">
    <keyboard-shortcut first-keystroke="meta X" />
  </action>
</keymap>`,
			validateFunc: func(t *testing.T, out KeymapXML) {
				assert.Len(t, out.Actions, 2)

				// Check managed action ($Copy)
				var copyAction *ActionXML
				var customAction *ActionXML
				for i := range out.Actions {
					switch out.Actions[i].ID {
					case "$Copy":
						copyAction = &out.Actions[i]
					case "CustomUserAction":
						customAction = &out.Actions[i]
					}
				}

				assert.NotNil(t, copyAction, "expected $Copy action")
				assert.NotNil(t, customAction, "expected CustomUserAction to be preserved")

				if copyAction != nil {
					assert.Len(t, copyAction.KeyboardShortcuts, 1)
					assert.Equal(t, "meta C", copyAction.KeyboardShortcuts[0].First)
				}
				if customAction != nil {
					assert.Len(t, customAction.KeyboardShortcuts, 1)
					assert.Equal(t, "meta X", customAction.KeyboardShortcuts[0].First)
				}
			},
		},
		{
			name: "managed action takes priority over conflicting user action",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<keymap version="1" name="Onekeymap" parent="$default" disable-mnemonics="true">
  <action id="$Copy">
    <keyboard-shortcut first-keystroke="meta V" />
  </action>
  <action id="CustomUserAction">
    <keyboard-shortcut first-keystroke="meta X" />
  </action>
</keymap>`,
			validateFunc: func(t *testing.T, out KeymapXML) {
				assert.Len(t, out.Actions, 2)

				var copyAction *ActionXML
				var customAction *ActionXML
				for i := range out.Actions {
					switch out.Actions[i].ID {
					case "$Copy":
						copyAction = &out.Actions[i]
					case "CustomUserAction":
						customAction = &out.Actions[i]
					}
				}

				assert.NotNil(t, copyAction, "expected $Copy action")
				assert.NotNil(t, customAction, "expected CustomUserAction to be preserved")

				// Managed action should override user's conflicting action
				if copyAction != nil {
					assert.Len(t, copyAction.KeyboardShortcuts, 1)
					assert.Equal(
						t,
						"meta C",
						copyAction.KeyboardShortcuts[0].First,
						"managed action should take priority",
					)
				}
			},
		},
		{
			name: "empty existing config behaves as destructive export",
			setting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<keymap version="1" name="Onekeymap" parent="$default" disable-mnemonics="true">
</keymap>`,
			validateFunc: func(t *testing.T, out KeymapXML) {
				assert.Len(t, out.Actions, 1)

				var copyAction *ActionXML
				for i := range out.Actions {
					if out.Actions[i].ID == "$Copy" {
						copyAction = &out.Actions[i]
						break
					}
				}

				assert.NotNil(t, copyAction, "expected $Copy action")
				if copyAction != nil {
					assert.Len(t, copyAction.KeyboardShortcuts, 1)
					assert.Equal(t, "meta C", copyAction.KeyboardShortcuts[0].First)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())
			exporter, err := plugin.Exporter()
			require.NoError(t, err)

			buf := &bytes.Buffer{}
			opts := pluginapi.PluginExportOption{}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			report, err := exporter.Export(context.TODO(), buf, tt.setting, opts)
			require.NoError(t, err)
			require.NotNil(t, report)

			var out KeymapXML
			require.NoError(t, xml.Unmarshal(buf.Bytes(), &out))

			tt.validateFunc(t, out)
		})
	}
}
