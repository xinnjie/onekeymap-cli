package xcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb, err := parseKeybinding(tt.input)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := formatKeybinding(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}
