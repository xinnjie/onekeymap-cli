package zed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/zed"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

func TestZed_FormatKeybinding(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "Simple", in: "ctrl+k", want: "ctrl-k"},
		{name: "Special Enter", in: "ctrl+enter", want: "ctrl-enter"},
		{name: "ManyModifiersFunction", in: "meta+ctrl+shift+alt+f5", want: "cmd-ctrl-shift-alt-f5"},
		{name: "MultiChord", in: "ctrl+k ctrl+s", want: "ctrl-k ctrl-s"},
		{name: "MetaAliasCmd", in: "cmd+k", want: "cmd-k"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := keybinding.NewKeybinding(tc.in, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
			require.NoError(t, err)
			out, err := zed.FormatZedKeybind(kb)
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestZed_ParseKeybinding(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "Simple", in: "ctrl-k", want: "ctrl-k"},
		{name: "Special Enter", in: "ctrl-enter", want: "ctrl-enter"},
		{name: "ManyModifiersFunction", in: "cmd-ctrl-shift-alt-f5", want: "meta-ctrl-shift-alt-f5"},
		{name: "MultiChord", in: "ctrl-k ctrl-s", want: "ctrl-k ctrl-s"},
		{name: "Empty", in: "", wantErr: true},
		{name: "UnknownKey", in: "ctrl-unknown", wantErr: true},
		// TODO(xinnjie): Not sure whether single shift is valid. But `shift shift` need to be valid
		{name: "ModifierOnly", in: "shift", want: "shift"},
		{name: "ModifierTwoShift", in: "shift shift", want: "shift shift"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := zed.ParseZedKeybind(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, kb.KeyChords)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, kb.KeyChords)
			// Normalize using Linux and '-' to canonical form for zed
			out := kb.String(keybinding.FormatOption{Platform: platform.PlatformLinux, Separator: "-"})
			assert.Equal(t, tc.want, out)
		})
	}
}
