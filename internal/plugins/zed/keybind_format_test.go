package zed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
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
			kb := keymap.MustParseKeyBinding(tc.in)
			out, err := formatZedKeybind(kb)
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
			kb, err := parseZedKeybind(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, kb)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, kb)
			// Normalize using Linux and '-' to canonical form for zed
			out, err := kb.Format(platform.PlatformLinux, "-")
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}
