package helix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
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
			kb, err := keybinding.NewKeybinding(tc.in, keybinding.ParseOption{Separator: "+"})
			require.NoError(t, err)
			out, err := formatKeybinding(kb)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tc.want, out)
		})
	}
}
