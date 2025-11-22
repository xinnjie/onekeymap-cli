package xcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

func TestParseKeybinding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{
			name:     "Cmd+K",
			input:    "@k",
			wantErr:  false,
			expected: "cmd+k",
		},
		{
			name:     "Ctrl+G",
			input:    "^g",
			wantErr:  false,
			expected: "ctrl+g",
		},
		{
			name:     "Cmd+Shift+J",
			input:    "@$j",
			wantErr:  false,
			expected: "cmd+shift+j",
		},
		{
			name:     "Alt+Tab",
			input:    "~\t",
			wantErr:  false,
			expected: "alt+tab",
		},
		{
			name:     "Ctrl+Up Arrow",
			input:    "^\uF700",
			wantErr:  false,
			expected: "ctrl+up",
		},
		{
			name:     "Option+Page Down",
			input:    "~\uF72D",
			wantErr:  false,
			expected: "alt+pagedown",
		},
		{
			name:     "Cmd+Shift+Down Arrow",
			input:    "@$\uF701",
			wantErr:  false,
			expected: "cmd+shift+down",
		},
		{
			name:     "F5",
			input:    "\uF708",
			wantErr:  false,
			expected: "f5",
		},
		{
			name:     "Cmd+Home",
			input:    "@\uF729",
			wantErr:  false,
			expected: "cmd+home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb, err := xcode.ParseKeybinding(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			actual := kb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFormatKeybinding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Cmd key",
			input:    "cmd+k",
			expected: "@k",
		},
		{
			name:     "Ctrl+G",
			input:    "ctrl+g",
			expected: "^g",
		},
		{
			name:     "Complex combination",
			input:    "alt+cmd+l",
			expected: "@~l",
		},
		{
			name:     "Page Down",
			input:    "pagedown",
			expected: "\uF72D",
		},
		{
			name:     "Ctrl+Up Arrow",
			input:    "ctrl+up",
			expected: "^\uF700",
		},
		{
			name:     "Cmd+F5",
			input:    "cmd+f5",
			expected: "@\uF708",
		},
		{
			name:     "shift+alt+left",
			input:    "shift+alt+left",
			expected: "$~\uf702",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb, err := keybinding.NewKeybinding(
				tt.input,
				keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"},
			)
			require.NoError(t, err)
			formatted, err := xcode.FormatKeybinding(kb)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}
