package vscode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

func TestVSCode_FormatKeybinding(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "Simple", in: "ctrl+k", want: "ctrl+k"},
		{name: "Special Enter", in: "ctrl+enter", want: "ctrl+enter"},
		{name: "ManyModifiersFunction", in: "meta+ctrl+shift+alt+f5", want: "cmd+ctrl+shift+alt+f5"},
		{name: "MultiChord", in: "ctrl+k ctrl+s", want: "ctrl+k ctrl+s"},
		{name: "MetaAliasCmd", in: "cmd+k", want: "cmd+k"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := keybinding.NewKeybinding(tc.in, keybinding.ParseOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			})
			require.NoError(t, err)
			out, err := vscode.FormatKeybinding(&kb, platform.PlatformMacOS)
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestVSCode_ParseKeybinding(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "Simple", in: "ctrl+k", want: "ctrl+k"},
		{name: "Special Enter", in: "ctrl+enter", want: "ctrl+enter"},
		{name: "ManyModifiersFunction", in: "cmd+ctrl+shift+alt+f5", want: "meta+ctrl+shift+alt+f5"},
		{name: "MultiChord", in: "ctrl+k ctrl+s", want: "ctrl+k ctrl+s"},
		{name: "Empty", in: "", wantErr: true},
		{name: "UnknownKey", in: "ctrl+unknown", wantErr: true},
		{name: "ModifierOnly", in: "shift", want: "shift"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := vscode.ParseKeybinding(tc.in, platform.PlatformMacOS)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, kb)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, kb)
			// Normalize using Linux and '+' to canonical form
			out := kb.String(keybinding.FormatOption{
				Platform:  platform.PlatformLinux,
				Separator: "+",
			})
			assert.Equal(t, tc.want, out)
		})
	}
}

// TestVSCode_CrossPlatformFormat tests that keybindings are formatted correctly
// for different target platforms, ensuring cross-platform compatibility.
func TestVSCode_CrossPlatformFormat(t *testing.T) {
	// Create a keybinding with meta modifier (internal representation)
	kb, err := keybinding.NewKeybinding("meta+k", keybinding.ParseOption{
		Platform:  platform.PlatformLinux, // Use Linux to get canonical "meta" form
		Separator: "+",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		platform platform.Platform
		want     string
	}{
		{name: "macOS uses cmd", platform: platform.PlatformMacOS, want: "cmd+k"},
		{name: "Windows uses win", platform: platform.PlatformWindows, want: "win+k"},
		{name: "Linux uses meta", platform: platform.PlatformLinux, want: "meta+k"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := vscode.FormatKeybinding(&kb, tc.platform)
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

// TestVSCode_CrossPlatformParse tests that keybindings with platform-specific
// modifier names are parsed correctly.
func TestVSCode_CrossPlatformParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		parsePlatform platform.Platform
		wantMeta      bool // Whether the parsed keybinding should have meta modifier
	}{
		{name: "macOS cmd parsed as meta", input: "cmd+k", parsePlatform: platform.PlatformMacOS, wantMeta: true},
		{name: "Windows win parsed as meta", input: "win+k", parsePlatform: platform.PlatformWindows, wantMeta: true},
		{name: "Linux meta parsed as meta", input: "meta+k", parsePlatform: platform.PlatformLinux, wantMeta: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := vscode.ParseKeybinding(tc.input, tc.parsePlatform)
			require.NoError(t, err)
			require.NotNil(t, kb)
			// Verify by formatting to Linux (canonical form)
			out := kb.String(keybinding.FormatOption{
				Platform:  platform.PlatformLinux,
				Separator: "+",
			})
			if tc.wantMeta {
				assert.Equal(t, "meta+k", out)
			}
		})
	}
}
