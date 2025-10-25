package cmd

import (
	"testing"

	ij "github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
)

func TestReplaceControlWithMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple control + key",
			input:    "control C",
			expected: "meta C",
		},
		{
			name:     "control with multiple modifiers",
			input:    "control alt S",
			expected: "meta alt S",
		},
		{
			name:     "control shift combo",
			input:    "control shift ENTER",
			expected: "meta shift ENTER",
		},
		{
			name:     "control at end",
			input:    "shift control",
			expected: "shift meta",
		},
		{
			name:     "uppercase CONTROL",
			input:    "CONTROL C",
			expected: "meta C",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no control modifier",
			input:    "alt S",
			expected: "alt S",
		},
		{
			name:     "meta already present",
			input:    "meta C",
			expected: "meta C",
		},
		{
			name:     "control with insert",
			input:    "control INSERT",
			expected: "meta INSERT",
		},
		{
			name:     "multiple words with control",
			input:    "control alt shift F7",
			expected: "meta alt shift F7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceControlWithMeta(tt.input)
			if result != tt.expected {
				t.Errorf("replaceControlWithMeta(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertControlToMetaForMac(t *testing.T) {
	doc := &ij.KeymapXML{
		Actions: []ij.ActionXML{
			{
				ID: "$Copy", // Inherited action - should be converted
				ShortcutXML: ij.ShortcutXML{
					KeyboardShortcuts: []ij.KeyboardShortcutXML{
						{First: "control C"},
						{First: "control INSERT"},
					},
				},
			},
			{
				ID: "EditorJoinLines", // Inherited action - should be converted
				ShortcutXML: ij.ShortcutXML{
					KeyboardShortcuts: []ij.KeyboardShortcutXML{
						{First: "control shift J"},
					},
				},
			},
			{
				ID: "CodeCompletion", // Explicit Mac action - should NOT be converted
				ShortcutXML: ij.ShortcutXML{
					KeyboardShortcuts: []ij.KeyboardShortcutXML{
						{First: "control SPACE"},
					},
				},
			},
			{
				ID: "NoControl", // Inherited action without control - unchanged
				ShortcutXML: ij.ShortcutXML{
					KeyboardShortcuts: []ij.KeyboardShortcutXML{
						{First: "alt F7"},
					},
				},
			},
		},
	}

	// Simulate Mac OS X.xml where CodeCompletion is explicitly defined
	meta := &keymapMetadata{
		topLevelActions: map[string]bool{
			"CodeCompletion": true, // Explicitly defined in Mac OS X.xml
		},
	}

	convertControlToMetaForMac(doc, meta)

	// Verify $Copy action (inherited, should be converted)
	if len(doc.Actions[0].KeyboardShortcuts) != 2 {
		t.Fatalf("expected 2 shortcuts for $Copy, got %d", len(doc.Actions[0].KeyboardShortcuts))
	}
	if doc.Actions[0].KeyboardShortcuts[0].First != "meta C" {
		t.Errorf("$Copy[0]: expected 'meta C', got %q", doc.Actions[0].KeyboardShortcuts[0].First)
	}
	if doc.Actions[0].KeyboardShortcuts[1].First != "meta INSERT" {
		t.Errorf("$Copy[1]: expected 'meta INSERT', got %q", doc.Actions[0].KeyboardShortcuts[1].First)
	}

	// Verify EditorJoinLines action (inherited, should be converted)
	if doc.Actions[1].KeyboardShortcuts[0].First != "meta shift J" {
		t.Errorf("EditorJoinLines: expected 'meta shift J', got %q", doc.Actions[1].KeyboardShortcuts[0].First)
	}

	// Verify CodeCompletion action (explicit Mac, should NOT be converted)
	if doc.Actions[2].KeyboardShortcuts[0].First != "control SPACE" {
		t.Errorf(
			"CodeCompletion: expected 'control SPACE' (preserved), got %q",
			doc.Actions[2].KeyboardShortcuts[0].First,
		)
	}

	// Verify NoControl action (inherited, no control to convert)
	if doc.Actions[3].KeyboardShortcuts[0].First != "alt F7" {
		t.Errorf("NoControl: expected 'alt F7', got %q", doc.Actions[3].KeyboardShortcuts[0].First)
	}
}
