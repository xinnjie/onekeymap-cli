package xcode

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

func TestXcodeKeybindingLookup(t *testing.T) {
	tests := []struct {
		name         string
		plistXML     string
		keybind      string
		expectedKeys []string
		wantErr      bool
	}{
		{
			name: "Find cmd+j keybinding",
			plistXML: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Menu Key Bindings</key>
  <dict>
    <key>Key Bindings</key>
    <array>
      <dict>
        <key>Action</key>
        <string>editorContext_jumpToDefinition:</string>
        <key>Alternate</key>
        <string>NO</string>
        <key>CommandID</key>
        <string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
        <key>Group</key>
        <string>Navigate Menu</string>
        <key>GroupID</key>
        <string>Xcode.IDEKit.MenuDefinition.Main</string>
        <key>GroupedAlternate</key>
        <string>NO</string>
        <key>Keyboard Shortcut</key>
        <string>@j</string>
        <key>Navigation</key>
        <string>YES</string>
        <key>Title</key>
        <string>Jump to Definition</string>
      </dict>
    </array>
  </dict>
</dict>
</plist>`,
			keybind: "cmd+j",
			expectedKeys: []string{
				"Action",
				"CommandID",
				"KeyboardShortcut",
				"Title",
				"Alternate",
				"Group",
				"GroupID",
				"GroupedAlternate",
				"Navigation",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := NewXcodeKeybindingLookup()

			keybinding, err := keybinding.NewKeybinding(tt.keybind, keybinding.ParseOption{Separator: "+"})
			require.NoError(t, err)

			results, err := lookup.Lookup(strings.NewReader(tt.plistXML), keybinding)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectedKeys == nil {
				assert.Empty(t, results)
				return
			}

			assert.NotEmpty(t, results)

			for _, result := range results {
				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(result), &parsed)
				require.NoError(t, err, "Result should be valid JSON: %s", result)

				for _, key := range tt.expectedKeys {
					assert.Contains(t, parsed, key, "Missing key %s in result: %s", key, result)
				}

				if keyboardShortcut, ok := parsed["KeyboardShortcut"].(string); ok {
					assert.NotEmpty(t, keyboardShortcut)
				}
			}
		})
	}
}
