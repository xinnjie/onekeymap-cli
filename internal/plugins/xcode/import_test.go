package xcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePlistKeybindings(t *testing.T) {
	tests := []struct {
		name     string
		xmlData  string
		expected xcodeKeybindingsPlist
		wantErr  bool
	}{
		{
			name: "Jump to Definition keybinding",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
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
			expected: xcodeKeybindingsPlist{
				MenuKeyBindings: menuKeyBindings{
					KeyBindings: []xcodeKeybinding{
						{
							Action:           "editorContext_jumpToDefinition:",
							CommandID:        "Xcode.IDEKit.CmdDefinition.JumpToDefinition",
							KeyboardShortcut: "@j",
							Title:            "Jump to Definition",
							Alternate:        "NO",
							Group:            "Navigate Menu",
							GroupID:          "Xcode.IDEKit.MenuDefinition.Main",
							GroupedAlternate: "NO",
							Navigation:       "YES",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple keybindings",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>clearConsole:</string>
				<key>CommandID</key>
				<string>Xcode.ConsoleKit.CmdDefinition.ClearConsole</string>
				<key>Keyboard Shortcut</key>
				<string>@k</string>
				<key>Title</key>
				<string>Clear Console</string>
			</dict>
			<dict>
				<key>Action</key>
				<string>moveCurrentLineUp:</string>
				<key>CommandID</key>
				<string>Xcode.IDEPegasusSourceEditor.CmdDefinition.MoveLineUp</string>
				<key>Keyboard Shortcut</key>
				<string>$~up</string>
				<key>Title</key>
				<string>Move Line Up</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expected: xcodeKeybindingsPlist{
				MenuKeyBindings: menuKeyBindings{
					KeyBindings: []xcodeKeybinding{
						{
							Action:           "clearConsole:",
							CommandID:        "Xcode.ConsoleKit.CmdDefinition.ClearConsole",
							KeyboardShortcut: "@k",
							Title:            "Clear Console",
						},
						{
							Action:           "moveCurrentLineUp:",
							CommandID:        "Xcode.IDEPegasusSourceEditor.CmdDefinition.MoveLineUp",
							KeyboardShortcut: "$~up",
							Title:            "Move Line Up",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty array",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
		</array>
	</dict>
</dict>
</plist>`,
			expected: xcodeKeybindingsPlist{
				MenuKeyBindings: menuKeyBindings{
					KeyBindings: []xcodeKeybinding{},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid XML",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<invalid>xml</invalid>`,
			expected: xcodeKeybindingsPlist{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plistData, err := parseXcodeConfig([]byte(tt.xmlData))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, *plistData)
		})
	}
}
