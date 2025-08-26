package intellij

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKeyStroke(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{"alt+HOME", "alt HOME", []string{"alt", "home"}, false},
		{"ctrl+alt+S (control alias)", "control alt S", []string{"ctrl", "alt", "s"}, false},
		{"shift+TAB", "shift TAB", []string{"shift", "tab"}, false},
		{"ctrl+shift+[", "control shift OPEN_BRACKET", []string{"ctrl", "shift", "["}, false},
		{"ctrl+shift+]", "control shift CLOSE_BRACKET", []string{"ctrl", "shift", "]"}, false},
		{"ctrl+`", "control BACK_QUOTE", []string{"ctrl", "`"}, false},
		{"ctrl+.", "control PERIOD", []string{"ctrl", "."}, false},
		{"ctrl+-", "control MINUS", []string{"ctrl", "-"}, false},
		{"ctrl+=", "control EQUALS", []string{"ctrl", "="}, false},
		{"ctrl+numpad_divide", "control DIVIDE", []string{"ctrl", "numpad_divide"}, false},
		{"ctrl+numpad3", "control NUMPAD3", []string{"ctrl", "numpad3"}, false},
		{"f-key", "F5", []string{"f5"}, false},
		{"unknown-key", "control UNKNOWN_KEY", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parts, err := parseKeyStroke(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, parts)
		})
	}
}

func TestFormatKeyStroke(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{"alt+HOME", []string{"alt", "home"}, "alt HOME"},
		{"ctrl+alt+S", []string{"ctrl", "alt", "s"}, "control alt S"},
		{"shift+TAB", []string{"shift", "tab"}, "shift TAB"},
		{"ctrl+shift+[", []string{"ctrl", "shift", "["}, "control shift OPEN_BRACKET"},
		{"ctrl+shift+]", []string{"ctrl", "shift", "]"}, "control shift CLOSE_BRACKET"},
		{"ctrl+`", []string{"ctrl", "`"}, "control BACK_QUOTE"},
		{"ctrl+.", []string{"ctrl", "."}, "control PERIOD"},
		{"ctrl+-", []string{"ctrl", "-"}, "control MINUS"},
		{"ctrl+=", []string{"ctrl", "="}, "control EQUALS"},
		{"ctrl+numpad_divide", []string{"ctrl", "numpad_divide"}, "control DIVIDE"},
		{"ctrl+numpad3", []string{"ctrl", "numpad3"}, "control NUMPAD3"},
		{"f-key", []string{"f5"}, "F5"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := formatKeyChord(tc.input)
			assert.Equal(t, tc.expected, out)
		})
	}
}
