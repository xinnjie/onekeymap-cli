package xcode

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func testMappingConfig() *mappings.MappingConfig {
	return &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.navigation.jumpToDefinition": {
				ID:          "actions.navigation.jumpToDefinition",
				Description: "Jump to definition",
				Xcode: []mappings.XcodeMappingConfig{
					{
						MenuAction: mappings.XcodeMenuAction{
							Action:    "editorContext_jumpToDefinition:",
							CommandID: "Xcode.IDEKit.CmdDefinition.JumpToDefinition",
						},
					},
				},
			},
			"actions.cursor.pageDown": {
				ID:          "actions.cursor.pageDown",
				Description: "Page down",
				Xcode: []mappings.XcodeMappingConfig{
					{
						TextAction: mappings.XcodeTextAction{
							TextAction: mappings.StringOrSlice{Items: []string{"pageDown:"}},
						},
					},
				},
			},
			"actions.edit.insertLineAfter": {
				ID:          "actions.edit.insertLineAfter",
				Description: "Insert line after",
				Xcode: []mappings.XcodeMappingConfig{
					{
						TextAction: mappings.XcodeTextAction{
							TextAction: mappings.StringOrSlice{Items: []string{"moveToEndOfLine:", "insertNewline:"}},
						},
					},
				},
			},
		},
	}
}

func TestExporter_SkipActions_MultipleKeybindings_Menu(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	// Same action with two bindings; Xcode supports only one
	a := keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j")
	a.Bindings = append(
		a.Bindings,
		&keymapv1.KeybindingReadable{KeyChords: keymap.MustParseKeyBinding("meta+k").KeyChords},
	)
	keymapSetting := &keymapv1.Keymap{Actions: []*keymapv1.Action{a}}

	var out bytes.Buffer
	report, err := exporter.Export(context.Background(), &out, keymapSetting, pluginapi.PluginExportOption{})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.GreaterOrEqual(t, len(report.SkipReport.SkipActions), 1)
	assert.Equal(t, "actions.navigation.jumpToDefinition", report.SkipReport.SkipActions[0].Action)
}

func TestExporter_SkipActions_MultipleKeybindings_Text(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	// Same action with two bindings; Xcode supports only one
	a := keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v")
	a.Bindings = append(
		a.Bindings,
		&keymapv1.KeybindingReadable{KeyChords: keymap.MustParseKeyBinding("ctrl+d").KeyChords},
	)
	keymapSetting := &keymapv1.Keymap{Actions: []*keymapv1.Action{a}}

	var out bytes.Buffer
	report, err := exporter.Export(context.Background(), &out, keymapSetting, pluginapi.PluginExportOption{})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.GreaterOrEqual(t, len(report.SkipReport.SkipActions), 1)
	assert.Equal(t, "actions.cursor.pageDown", report.SkipReport.SkipActions[0].Action)
}

func TestExporter_SkipActions_UnsupportedAction(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	// Unknown action not present in mapping config
	keymapSetting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.unknown.notMapped", "meta+u"),
		},
	}

	var out bytes.Buffer
	report, err := exporter.Export(context.Background(), &out, keymapSetting, pluginapi.PluginExportOption{})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.GreaterOrEqual(t, len(report.SkipReport.SkipActions), 1)
	assert.Equal(t, "actions.unknown.notMapped", report.SkipReport.SkipActions[0].Action)
}

func TestExporter_Export_MenuKeyBindings(t *testing.T) {
	mappingConfig := testMappingConfig()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		existingConfig string
		expectedConfig string
	}{
		{
			name: "exports basic menu key binding",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "non-destructive merge preserves unmanaged keybindings",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>customUserAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@k</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
				<dict>
					<key>Action</key>
					<string>customUserAction:</string>
					<key>Alternate</key>
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string/>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@k</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "managed keybinding takes priority over conflicting keybinding",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>conflictingAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@j</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var existingReader io.Reader
			if tt.existingConfig != "" {
				existingReader = strings.NewReader(tt.existingConfig)
			}

			_, err := exporter.Export(
				context.Background(),
				&out,
				tt.keymapSetting,
				pluginapi.PluginExportOption{ExistingConfig: existingReader},
			)
			require.NoError(t, err)

			// Compare with expected config string
			actual := out.String()
			assert.Equal(t, tt.expectedConfig, actual)
		})
	}
}

func TestExporter_Export_TextKeyBindings(t *testing.T) {
	mappingConfig := testMappingConfig()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		existingConfig string
		expectedConfig string
	}{
		{
			name: "exports text key binding",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v"),
				},
			},
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Menu Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<array>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
				<key>^v</key>
				<string>pageDown:</string>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "text bindings merge with existing preserving unmanaged",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
			<key>^d</key>
			<string>customTextAction:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Menu Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<array>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
				<key>^d</key>
				<string>customTextAction:</string>
				<key>^v</key>
				<string>pageDown:</string>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "managed text binding overrides conflicting existing",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.cursor.pageDown", "ctrl+v"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
			<string>oldAction:</string>
		</dict>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Menu Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<array>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
				<key>^v</key>
				<string>pageDown:</string>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "exports array text actions",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.insertLineAfter", "meta+d"),
				},
			},
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Menu Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<array>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
				<key>@d</key>
				<array>
					<string>moveToEndOfLine:</string>
					<string>insertNewline:</string>
				</array>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var existingReader io.Reader
			if tt.existingConfig != "" {
				existingReader = strings.NewReader(tt.existingConfig)
			}

			_, err := exporter.Export(
				context.Background(),
				&out,
				tt.keymapSetting,
				pluginapi.PluginExportOption{ExistingConfig: existingReader},
			)
			require.NoError(t, err)

			// Compare with expected config string
			actual := out.String()
			assert.Equal(t, tt.expectedConfig, actual)
		})
	}
}

func TestExporter_OrderByBaseCommand(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		existingConfig string
		expectedConfig string
	}{
		{
			name: "reorders according to base config",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>customUserAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@k</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
				<dict>
					<key>Action</key>
					<string>customUserAction:</string>
					<key>Alternate</key>
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string/>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@k</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var existingReader io.Reader
			if tt.existingConfig != "" {
				existingReader = strings.NewReader(tt.existingConfig)
			}

			_, err := exporter.Export(
				context.Background(),
				&out,
				tt.keymapSetting,
				pluginapi.PluginExportOption{ExistingConfig: existingReader},
			)
			require.NoError(t, err)

			// Compare with expected config string
			actual := out.String()
			assert.Equal(t, tt.expectedConfig, actual)
		})
	}
}

func TestExporter_MergeKeybindings(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		existingConfig string
		expectedConfig string
	}{
		{
			name: "merges managed and unmanaged without conflicts",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>customUserAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@u</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
				<dict>
					<key>Action</key>
					<string>customUserAction:</string>
					<key>Alternate</key>
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string/>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@u</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
		{
			name: "managed takes priority on keyboard shortcut conflict",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Menu Key Bindings</key>
	<dict>
		<key>Key Bindings</key>
		<array>
			<dict>
				<key>Action</key>
				<string>conflictingAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@j</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var existingReader io.Reader
			if tt.existingConfig != "" {
				existingReader = strings.NewReader(tt.existingConfig)
			}

			_, err := exporter.Export(
				context.Background(),
				&out,
				tt.keymapSetting,
				pluginapi.PluginExportOption{ExistingConfig: existingReader},
			)
			require.NoError(t, err)

			// Compare with expected config string
			actual := out.String()
			assert.Equal(t, tt.expectedConfig, actual)
		})
	}
}

func TestExporter_IdentifyUnmanagedKeybindings(t *testing.T) {
	mappingConfig := testMappingConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer())

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		existingConfig string
		expectedConfig string
	}{
		{
			name: "identifies and preserves only unmanaged keybindings",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.navigation.jumpToDefinition", "meta+j"),
				},
			},
			existingConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
				<string>@old</string>
			</dict>
			<dict>
				<key>Action</key>
				<string>customUserAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@c1</string>
			</dict>
			<dict>
				<key>Action</key>
				<string>anotherCustomAction:</string>
				<key>Keyboard Shortcut</key>
				<string>@c2</string>
			</dict>
		</array>
	</dict>
</dict>
</plist>`,
			expectedConfig: `<?xml version="1.0" encoding="UTF-8"?>
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
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string>Xcode.IDEKit.CmdDefinition.JumpToDefinition</string>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@j</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
				<dict>
					<key>Action</key>
					<string>customUserAction:</string>
					<key>Alternate</key>
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string/>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@c1</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
				<dict>
					<key>Action</key>
					<string>anotherCustomAction:</string>
					<key>Alternate</key>
					<string/>
					<key>CommandGroupID</key>
					<string/>
					<key>CommandID</key>
					<string/>
					<key>Group</key>
					<string/>
					<key>GroupID</key>
					<string/>
					<key>GroupedAlternate</key>
					<string/>
					<key>Keyboard Shortcut</key>
					<string>@c2</string>
					<key>Navigation</key>
					<string/>
					<key>Parent Title</key>
					<string/>
					<key>Title</key>
					<string/>
				</dict>
			</array>
			<key>Version</key>
			<integer>3</integer>
		</dict>
		<key>Text Key Bindings</key>
		<dict>
			<key>Key Bindings</key>
			<dict>
			</dict>
			<key>Version</key>
			<integer>3</integer>
		</dict>
	</dict>
</plist>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var existingReader io.Reader
			if tt.existingConfig != "" {
				existingReader = strings.NewReader(tt.existingConfig)
			}

			_, err := exporter.Export(
				context.Background(),
				&out,
				tt.keymapSetting,
				pluginapi.PluginExportOption{ExistingConfig: existingReader},
			)
			require.NoError(t, err)

			// Compare with expected config string
			actual := out.String()
			assert.Equal(t, tt.expectedConfig, actual)
		})
	}
}
