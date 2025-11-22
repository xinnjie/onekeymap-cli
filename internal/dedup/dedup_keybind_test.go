package dedup_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/onekeymap-cli/internal/dedup"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
)

func newAction(name string, keys ...string) keymap.Action {
	var bindings []keybinding.Keybinding
	for _, k := range keys {
		kb, err := keybinding.NewKeybinding(k, keybinding.ParseOption{
			Separator: "+",
			Platform:  platform.PlatformMacOS,
		})
		if err != nil {
			panic(err)
		}
		bindings = append(bindings, kb)
	}
	return keymap.Action{
		Name:     name,
		Bindings: bindings,
	}
}

func TestActions_Table(t *testing.T) {
	tests := []struct {
		name     string
		input    []keymap.Action
		expected []keymap.Action
	}{
		{
			name: "EmptyChords",
			input: []keymap.Action{
				{
					Name: "actions.copy",
					Bindings: []keybinding.Keybinding{
						{KeyChords: []keychord.KeyChord{}},
					},
				},
			},
			expected: []keymap.Action{},
		},
		{
			name: "DuplicatesByActionAndChords",
			input: []keymap.Action{
				newAction("actions.copy", "k"),
				newAction("actions.copy", "k"),
			},
			expected: []keymap.Action{
				newAction("actions.copy", "k"),
			},
		},
		{
			name: "KeepsOrderAndSkipsNil",
			input: []keymap.Action{
				newAction("actions.open", "o"),
				newAction("actions.open", "o"),
				newAction("actions.save", "o"),
			},
			expected: []keymap.Action{
				newAction("actions.open", "o"),
				newAction("actions.save", "o"),
			},
		},
		{
			name: "AggregatesByIDForDifferentChords",
			input: []keymap.Action{
				newAction("actions.find", "f"),
				newAction("actions.replace", "f", "f", "f"),
				newAction("actions.find", "g"),
			},
			expected: []keymap.Action{
				newAction("actions.find", "f", "g"),
				newAction("actions.replace", "f"),
			},
		},
		{
			name: "NilVsEmptyChordsConsideredEqual",
			input: []keymap.Action{
				{Name: "actions.nochord"},
				{Name: "actions.nochord", Bindings: []keybinding.Keybinding{}},
			},
			expected: []keymap.Action{
				{Name: "actions.nochord"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := dedup.Actions(tc.input)
			diff := cmp.Diff(tc.expected, actual)
			assert.Empty(t, diff)
		})
	}
}
