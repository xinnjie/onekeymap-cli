package helix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
)

func TestFormatKeybinding(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "Simple", in: "ctrl+k", want: "C-k"},
		{name: "Special Enter", in: "ctrl+enter", want: "C-ret"},
		{name: "ManyModifiersFunction", in: "meta+ctrl+shift+alt+f5", want: "M-C-S-A-F5"},
		{name: "MultiChord", in: "ctrl+k ctrl+s", want: "C-k C-s"},
		{name: "Minus", in: "ctrl+-", want: "C--"},
		{name: "Numpad", in: "ctrl+numpad0", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb := keymap.MustParseKeyBinding(tc.in)
			out, err := formatKeybinding(kb)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestParseKeybinding(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantNorm string
		wantErr  bool
	}{
		{name: "Simple", in: "C-k", wantNorm: "ctrl+k"},
		{name: "Special Enter", in: "C-ret", wantNorm: "ctrl+enter"},
		{name: "ManyModifiersFunction", in: "M-C-S-A-F5", wantNorm: "meta+ctrl+shift+alt+f5"},
		{name: "MultiChord", in: "C-k C-s", wantNorm: "ctrl+k ctrl+s"},
		{name: "Empty", in: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := parseKeybinding(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			norm, err := kb.Format(platform.PlatformLinux, "+")
			require.NoError(t, err)
			assert.Equal(t, tc.wantNorm, norm)
		})
	}
}
