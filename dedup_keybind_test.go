package onekeymap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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
					Bindings: []*keymapv1.Binding{
						{KeyChords: &keymapv1.KeyChordSequence{Chords: []*keymapv1.KeyChord{}}},
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
				{Name: "actions.nochord", Bindings: []*keymapv1.Binding{}},
			},
			expected: []*keymapv1.Action{
				{Name: "actions.nochord"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := dedupKeyBindings(tc.input)
			diff := cmp.Diff(tc.expected, actual, protocmp.Transform())
			assert.Empty(t, diff)
		})
	}
}
