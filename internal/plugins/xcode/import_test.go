package xcode

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/proto"
)

func TestImporter_Import(t *testing.T) {
	// Create test mapping config
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	importer := newImporter(mappingConfig, logger, metrics.NewNoop())

	tests := []struct {
		name     string
		xmlData  string
		expected *keymapv1.Keymap
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
				<key>CommandID</key>
				<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
				<key>Keyboard Shortcut</key>
				<string>@j</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "cmd+j"),
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple keybindings from same file",
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
				<key>CommandID</key>
				<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
				<key>Keyboard Shortcut</key>
				<string>@j</string>
			</dict>
		</array>
	</dict>
	<key>Text Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<dict>
			<key>^v</key>
			<string>pageDown:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "cmd+j"),
					keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v"),
				},
			},
			wantErr: false,
		},
		{
			name: "Empty keybindings",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array></array>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{},
			wantErr:  false,
		},
		{
			name: "Text Key Bindings single action but mapping is multi-item (skipped)",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array/>
	</dict>
	<key>Text Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<dict>
			<key>@d</key>
			<string>moveToEndOfLine:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{},
			wantErr:  false,
		},
		{
			name: "Text Key Bindings with single action",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array/>
	</dict>
	<key>Text Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<dict>
			<key>^v</key>
			<string>pageDown:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v"),
				},
			},
			wantErr: false,
		},
		{
			name: "Text Key Bindings with array of actions (skipped)",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array/>
	</dict>
	<key>Text Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<dict>
			<key>^$J</key>
			<array>
				<string>moveBackward:</string>
				<string>moveToBeginningOfParagraph:</string>
			</array>
		</dict>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{},
			wantErr:  false,
		},
		{
			name: "Text Key Bindings with unprintable Unicode characters (arrow keys, function keys)",
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
				<key>CommandID</key>
				<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
				<key>Keyboard Shortcut</key>
				<string>$@` + "\uF700" + `</string>
			</dict>
		</array>
	</dict>
	<key>Text Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<dict>
			<key>~` + "\uF72D" + `</key>
			<string>pageDown:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expected: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "shift+cmd+up"),
					keymap.NewActioinBinding("actions.cursor.pageDown", "alt+pagedown"),
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid XML",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<invalid>xml</invalid>`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xmlData)
			result, err := importer.Import(context.Background(), reader, pluginapi.PluginImportOption{})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, proto.Equal(tt.expected, result.Keymap), "Expected %v, got %v", tt.expected, result.Keymap)
		})
	}
}
