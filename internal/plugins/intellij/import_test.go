package intellij

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

func TestImportIntelliJKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	if err != nil {
		t.Fatal(err)
	}
	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformLinux, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	testCases := []struct {
		name      string
		input     string
		expected  keymap.Keymap
		expectErr bool
	}{
		{
			name: "Basic single-chord mapping ($Copy)",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="$Copy">
    <keyboard-shortcut first-keystroke="meta C"/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Multi-chord, well-known keys, and unmapped skip",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="control alt S"/>
    <keyboard-shortcut first-keystroke="control K" second-keystroke="control C"/>
  </action>
  <action id="UnmappedAction">
    <keyboard-shortcut first-keystroke="meta X"/>
  </action>
  <action id="TestAction">
    <keyboard-shortcut first-keystroke="shift HOME"/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s"), parseKB("ctrl+k ctrl+c")},
					},
					{
						Name:     "actions.test.withArgs",
						Bindings: []keybinding.Keybinding{parseKB("shift+home")},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Deduplicate identical keybindings",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="control alt S"/>
    <keyboard-shortcut first-keystroke="control alt S"/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s")},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid keystroke should be skipped (result empty)",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="alt UNKNOWN"/>
  </action>
</keymap>`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name: "Empty second-keystroke is accepted",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="control K" second-keystroke=""/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+k")},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Whitespace second-keystroke is invalid and skipped",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="control K" second-keystroke="  "/>
  </action>
</keymap>`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name: "Special keys: F-keys, numpad, brackets",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="F5"/>
    <keyboard-shortcut first-keystroke="control NUMPAD3"/>
    <keyboard-shortcut first-keystroke="control shift OPEN_BRACKET"/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{
							parseKB("f5"),
							parseKB("ctrl+numpad3"),
							parseKB("ctrl+shift+["),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "CTRL alias and mixed-case tokens are normalized",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="CTRL ALt s"/>
  </action>
</keymap>`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("ctrl+alt+s")},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			importer, err := plugin.Importer()
			require.NoError(t, err)
			result, err := importer.Import(context.Background(), reader, pluginapi.PluginImportOption{})

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Truef(t, reflect.DeepEqual(tc.expected, result.Keymap), "Expected and actual KeymapSetting should be equal, expect %v, got %v", tc.expected, result.Keymap)
			}
		})
	}
}
