package intellij

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestParseKeyBinding_Table(t *testing.T) {
	tests := []struct {
		name    string
		in      KeyboardShortcutXML
		want    string
		wantErr bool
	}{
		{name: "SingleChord", in: KeyboardShortcutXML{First: "alt HOME"}, want: "alt+home"},
		{name: "TwoChords", in: KeyboardShortcutXML{First: "control E", Second: "control S"}, want: "ctrl+e ctrl+s"},
		{name: "InvalidFirst", in: KeyboardShortcutXML{First: "control UNKNOWN_KEY"}, wantErr: true},
		{name: "InvalidSecond", in: KeyboardShortcutXML{First: "alt HOME", Second: "control UNKNOWN_KEY"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := parseKeyBinding(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, kb)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, kb)
			out, err := kb.Format(platform.PlatformLinux, "+")
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func TestFormatKeybinding_Table(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		build      func() *keymap.KeyBinding
		wantFirst  string
		wantSecond string
		wantErr    bool
		wantNil    bool
	}{
		{name: "SingleChord", in: "ctrl+alt+s", wantFirst: "control alt S", wantSecond: ""},
		{name: "TwoChords", in: "ctrl+e ctrl+s", wantFirst: "control E", wantSecond: "control S"},
		{name: "TooManyChords", in: "ctrl+a ctrl+b ctrl+c", wantErr: true, wantNil: true},
		{name: "InvalidChordProto", build: func() *keymap.KeyBinding {
			return keymap.NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{Chords: []*keymapv1.KeyChord{{}}}})
		}, wantErr: true, wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var kb *keymap.KeyBinding
			if tc.build != nil {
				kb = tc.build()
			} else {
				kb = keymap.MustParseKeyBinding(tc.in)
			}
			ks, err := formatKeybinding(kb)
			if tc.wantErr {
				assert.Error(t, err)
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
