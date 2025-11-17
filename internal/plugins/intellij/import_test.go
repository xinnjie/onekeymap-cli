package intellij

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/proto"
)

func TestImportIntelliJKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	if err != nil {
		t.Fatal(err)
	}
	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())

	testCases := []struct {
		name      string
		input     string
		expected  *keymapv1.Keymap
		expectErr bool
	}{
		{
			name: "Basic single-chord mapping ($Copy)",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="$Copy">
    <keyboard-shortcut first-keystroke="meta C"/>
  </action>
</keymap>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
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
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.test.mutipleActions", "ctrl+alt+s", "ctrl+k ctrl+c"),
					keymap.NewActioinBinding("actions.test.withArgs", "shift+home"),
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
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.test.mutipleActions", "ctrl+alt+s"),
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
			expected:  &keymapv1.Keymap{},
			expectErr: false,
		},
		{
			name: "Empty second-keystroke is accepted",
			input: `<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="command1">
    <keyboard-shortcut first-keystroke="control K" second-keystroke=""/>
  </action>
</keymap>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.test.mutipleActions", "ctrl+k"),
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
			expected:  &keymapv1.Keymap{},
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
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.test.mutipleActions", "f5", "ctrl+numpad3", "ctrl+shift+["),
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
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.test.mutipleActions", "ctrl+alt+s"),
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
				assert.Truef(t, proto.Equal(tc.expected, result.Keymap), "Expected and actual KeymapSetting should be equal, expect %s, got %s", tc.expected.String(), result.Keymap.String())
			}
		})
	}
}
