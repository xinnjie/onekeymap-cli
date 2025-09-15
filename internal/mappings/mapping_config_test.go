package mappings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	actionmappings "github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings/action_mappings"
	"gopkg.in/yaml.v3"
)

func TestLoader_Load(t *testing.T) {
	t.Run("successfully loads both single and multiple vscode configs", func(t *testing.T) {
		content := `
mappings:
  - id: "action.single"
    vscode:
      command: "cmd.single"
      when: "ctx.single"
  - id: "action.toggle"
    vscode:
      - command: "cmd.toggle.show"
        when: "ctx.toggle.show"
      - command: "cmd.toggle.hide"
        when: "ctx.toggle.hide"
`
		reader := strings.NewReader(content)
		mappingData, err := load(reader)
		require.NoError(t, err)

		require.Len(t, mappingData.Mappings, 2)

		// Test single config
		singleMapping, ok := mappingData.Mappings["action.single"]
		require.True(t, ok)
		require.Len(t, singleMapping.VSCode, 1)
		assert.Equal(t, "cmd.single", singleMapping.VSCode[0].Command)
		assert.Equal(t, "ctx.single", singleMapping.VSCode[0].When)

		// Test toggle config
		toggleMapping, ok := mappingData.Mappings["action.toggle"]
		require.True(t, ok)
		require.Len(t, toggleMapping.VSCode, 2)
		assert.Equal(t, "cmd.toggle.show", toggleMapping.VSCode[0].Command)
		assert.Equal(t, "ctx.toggle.show", toggleMapping.VSCode[0].When)
		assert.Equal(t, "cmd.toggle.hide", toggleMapping.VSCode[1].Command)
		assert.Equal(t, "ctx.toggle.hide", toggleMapping.VSCode[1].When)
	})

	t.Run("returns error on duplicate vscode config", func(t *testing.T) {
		content := `
mappings:
  - id: "action.one"
    vscode:
      command: "cmd.one"
      when: "ctx.one"
---
mappings:
  - id: "action.two"
    vscode:
      command: "cmd.one"
      when: "ctx.one"
`
		reader := strings.NewReader(content)
		_, err := load(reader)
		require.Error(t, err)
		var derr *DuplicateActionMappingError
		require.ErrorAs(t, err, &derr, "expected DuplicateActionMappingError")
		assert.Equal(t, "vscode", derr.Editor)
		assert.NotEmpty(t, derr.Duplicates)
	})

	t.Run("returns error on duplicate vscode config between single and list", func(t *testing.T) {
		content := `
mappings:
  - id: "action.one"
    vscode:
      command: "cmd.one"
      when: "ctx.one"
---
mappings:
  - id: "action.two"
    vscode:
      - command: "cmd.one"
        when: "ctx.one"
`
		reader := strings.NewReader(content)
		_, err := load(reader)
		require.Error(t, err)
		var derr *DuplicateActionMappingError
		require.ErrorAs(t, err, &derr, "expected DuplicateActionMappingError")
		assert.Equal(t, "vscode", derr.Editor)
		assert.NotEmpty(t, derr.Duplicates)
	})

	t.Run("returns error on duplicate zed config", func(t *testing.T) {
		content := `
mappings:
  - id: "action.one"
    zed:
      action: "action.one"
      context: "ctx.one"
---
mappings:
  - id: "action.two"
    zed:
      action: "action.one"
      context: "ctx.one"
`
		reader := strings.NewReader(content)
		_, err := load(reader)
		require.Error(t, err)
		var derr *DuplicateActionMappingError
		require.ErrorAs(t, err, &derr, "expected DuplicateActionMappingError")
		assert.Equal(t, "zed", derr.Editor)
		assert.NotEmpty(t, derr.Duplicates)
	})

	t.Run("successfully loads and merges mappings from a multi-document stream", func(t *testing.T) {
		content := `
mappings:
  - id: "one.action.copy"
    description: "Copy"
    vscode:
      command: "editor.action.clipboardCopyAction"
---
mappings:
  - id: "one.action.paste"
    description: "Paste"
    vscode:
      command: "editor.action.clipboardPasteAction"
`
		reader := strings.NewReader(content)

		mappingData, err := load(reader)
		require.NoError(t, err)

		assert.Len(t, mappingData.Mappings, 2)
		assert.Equal(t, "Copy", mappingData.Mappings["one.action.copy"].Description)
		assert.Equal(t, "Paste", mappingData.Mappings["one.action.paste"].Description)
	})

	t.Run("returns error on duplicate action ID in stream", func(t *testing.T) {
		content := `
mappings:
  - id: "one.action.copy"
---
mappings:
  - id: "one.action.copy"
`
		reader := strings.NewReader(content)

		_, err := load(reader)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate action ID 'one.action.copy'")
	})

	t.Run("returns error on malformed YAML in stream", func(t *testing.T) {
		content := "mappings: [id: 'bad'"
		reader := strings.NewReader(content)

		_, err := load(reader)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML stream")
	})

	t.Run("successfully loads an empty stream", func(t *testing.T) {
		reader := strings.NewReader("")
		mappingData, err := load(reader)
		require.NoError(t, err)
		assert.NotNil(t, mappingData)
		assert.Empty(t, mappingData.Mappings)
	})
}

func TestNewMappingConfig_NoDuplicateIDs(t *testing.T) {
	// Validate that production mapping files contain no duplicate action IDs.
	reader, err := actionmappings.ReadActionMapping()
	require.NoError(t, err)
	dec := yaml.NewDecoder(reader)
	seen := make(map[string]struct{})
	for {
		var doc configFormat
		if derr := dec.Decode(&doc); derr != nil {
			if derr.Error() == "EOF" {
				break
			}
			// if decoding fails, fail fast
			require.NoError(t, derr)
		}
		for _, m := range doc.Mappings {
			if _, ok := seen[m.ID]; ok {
				t.Fatalf("duplicate action ID '%s' found in production configuration", m.ID)
			}
			seen[m.ID] = struct{}{}
		}
	}
}

func TestDuplicateInProductionActionMappingConfig(t *testing.T) {
	_, err := NewMappingConfig()
	require.NoError(t, err)
}
