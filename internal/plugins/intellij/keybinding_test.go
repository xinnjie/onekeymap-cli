package intellij_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

func TestParseKeyBinding_Table(t *testing.T) {
	tests := []struct {
		name    string
		in      intellij.KeyboardShortcutXML
		want    string
		wantErr bool
	}{
		{name: "SingleChord", in: intellij.KeyboardShortcutXML{First: "alt HOME"}, want: "alt+home"},
		{
			name: "TwoChords",
			in:   intellij.KeyboardShortcutXML{First: "control E", Second: "control S"},
			want: "ctrl+e ctrl+s",
		},
		{name: "InvalidFirst", in: intellij.KeyboardShortcutXML{First: "control UNKNOWN_KEY"}, wantErr: true},
		{
			name:    "InvalidSecond",
			in:      intellij.KeyboardShortcutXML{First: "alt HOME", Second: "control UNKNOWN_KEY"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := intellij.ParseKeyBinding(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			out := kb.String(keybinding.FormatOption{
				Platform:  platform.PlatformLinux,
				Separator: "+",
			})
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestFormatKeybinding_Table(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantFirst  string
		wantSecond string
		wantErr    bool
	}{
		{name: "SingleChord", in: "ctrl+alt+s", wantFirst: "control alt S", wantSecond: ""},
		{name: "TwoChords", in: "ctrl+e ctrl+s", wantFirst: "control E", wantSecond: "control S"},
		{name: "TooManyChords", in: "ctrl+a ctrl+b ctrl+c", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := keybinding.NewKeybinding(tc.in, keybinding.ParseOption{
				Platform:  platform.PlatformLinux,
				Separator: "+",
			})
			require.NoError(t, err)
			ks, err := intellij.FormatKeybinding(kb)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, ks)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ks)
			assert.Equal(t, tc.wantFirst, ks.First)
			assert.Equal(t, tc.wantSecond, ks.Second)
		})
	}
}
