package xcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
)

func TestParseKeybinding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected *keymap.KeyBinding
	}{
		{
			name:     "Cmd+K",
			input:    "@k",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("cmd+k"),
		},
		{
			name:     "Ctrl+G",
			input:    "^g",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("ctrl+g"),
		},
		{
			name:     "Cmd+Shift+J",
			input:    "@$j",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("cmd+shift+j"),
		},
		{
			name:     "Alt+Tab",
			input:    "~\t",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("alt+tab"),
		},
		{
			name:     "Ctrl+Up Arrow",
			input:    "^\uF700",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("ctrl+up"),
		},
		{
			name:     "Option+Page Down",
			input:    "~\uF72D",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("alt+pagedown"),
		},
		{
			name:     "Cmd+Shift+Down Arrow",
			input:    "@$\uF701",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("cmd+shift+down"),
		},
		{
			name:     "F5",
			input:    "\uF708",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("f5"),
		},
		{
			name:     "Cmd+Home",
			input:    "@\uF729",
			wantErr:  false,
			expected: keymap.MustParseKeyBinding("cmd+home"),
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

			assert.Equal(t, tt.expected, kb)
		})
	}
}

func TestFormatKeybinding(t *testing.T) {
	tests := []struct {
		name     string
		input    *keymap.KeyBinding
		expected string
	}{
		{
			name:     "Simple Cmd key",
			input:    keymap.MustParseKeyBinding("cmd+k"),
			expected: "@k",
		},
		{
			name:     "Ctrl+G",
			input:    keymap.MustParseKeyBinding("ctrl+g"),
			expected: "^g",
		},
		{
			name:     "Complex combination",
			input:    keymap.MustParseKeyBinding("alt+cmd+l"),
			expected: "@~l",
		},
		{
			name:     "Page Down",
			input:    keymap.MustParseKeyBinding("pagedown"),
			expected: "\uF72D",
		},
		{
			name:     "Ctrl+Up Arrow",
			input:    keymap.MustParseKeyBinding("ctrl+up"),
			expected: "^\uF700",
		},
		{
			name:     "Cmd+F5",
			input:    keymap.MustParseKeyBinding("cmd+f5"),
			expected: "@\uF708",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := xcode.FormatKeybinding(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}
