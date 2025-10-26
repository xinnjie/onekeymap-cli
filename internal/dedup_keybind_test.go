package internal_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestDedupKeyBindings_Table(t *testing.T) {
	tests := []struct {
		name     string
		input    []*keymapv1.Action
		expected []*keymapv1.Action
	}{
		{
			name: "EmptyChords",
			input: []*keymapv1.Action{
				{
					Name: "actions.copy",
					Bindings: []*keymapv1.KeybindingReadable{
						{KeyChords: &keymapv1.Keybinding{Chords: []*keymapv1.KeyChord{}}},
					},
				},
			},
			expected: []*keymapv1.Action{},
		},
		{
			name: "DuplicatesByActionAndChords",
			input: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.copy", "k"),
				keymap.NewActioinBinding("actions.copy", "k"),
			},
			expected: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.copy", "k"),
			},
		},
		{
			name: "KeepsOrderAndSkipsNil",
			input: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.open", "o"),
				keymap.NewActioinBinding("actions.open", "o"),
				keymap.NewActioinBinding("actions.save", "o"),
			},
			expected: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.open", "o"),
				keymap.NewActioinBinding("actions.save", "o"),
			},
		},
		{
			name: "AggregatesByIDForDifferentChords",
			input: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.find", "f"),
				keymap.NewActioinBinding("actions.replace", "f", "f", "f"),
				keymap.NewActioinBinding("actions.find", "g"),
			},
			expected: []*keymapv1.Action{
				keymap.NewActioinBinding("actions.find", "f", "g"),
				keymap.NewActioinBinding("actions.replace", "f"),
			},
		},
		{
			name: "NilVsEmptyChordsConsideredEqual",
			input: []*keymapv1.Action{
				{Name: "actions.nochord"},
				{Name: "actions.nochord", Bindings: []*keymapv1.KeybindingReadable{}},
			},
			expected: []*keymapv1.Action{
				{Name: "actions.nochord"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := internal.DedupKeyBindings(tc.input)
			diff := cmp.Diff(tc.expected, actual, protocmp.Transform())
			assert.Empty(t, diff)
		})
	}
}
