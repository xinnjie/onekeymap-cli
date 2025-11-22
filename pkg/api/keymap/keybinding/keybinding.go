package keybinding

import (
	"strings"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

type Keybinding struct {
	KeyChords []keychord.KeyChord
}

type ParseOption struct {
	Platform  platform.Platform
	Separator string
}

func NewKeybinding(keybinding string, opt ParseOption) (Keybinding, error) {
	parts := strings.Split(keybinding, " ")
	var chords []keychord.KeyChord
	for _, part := range parts {
		kc, err := keychord.NewKeyChord(part, keychord.ParseOption{
			Platform:  opt.Platform,
			Separator: opt.Separator,
		})
		if err != nil {
			return Keybinding{}, err
		}
		chords = append(chords, kc)
	}
	return Keybinding{KeyChords: chords}, nil
}

type FormatOption struct {
	Platform  platform.Platform
	Separator string
}

// String returns the string representation of the keybinding. like "ctrl+k ctrl+s"
func (kb Keybinding) String(opt FormatOption) string {
	var parts []string
	for _, chord := range kb.KeyChords {
		parts = append(parts, chord.String(keychord.FormatOption{
			Platform:  opt.Platform,
			Separator: opt.Separator,
		}))
	}
	return strings.Join(parts, " ")
}
