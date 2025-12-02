package mappings //nolint:testpackage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	actionmappings "github.com/xinnjie/onekeymap-cli/config/action_mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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

func TestGetExportAction(t *testing.T) {
	t.Run("returns action itself when supported", func(t *testing.T) {
		content := `
mappings:
  - id: "parent.action"
    children:
      - "child.action"
    vscode:
      command: "parent.cmd"
  - id: "child.action"
    vscode:
      command: "child.cmd"
`
		reader := strings.NewReader(content)
		mc, err := load(reader)
		require.NoError(t, err)

		action, fallback := mc.GetExportAction("parent.action", pluginapi.EditorTypeVSCode)
		require.NotNil(t, action)
		assert.Equal(t, "parent.action", action.ID)
		assert.False(t, fallback)
	})

	t.Run("falls back to child when parent not supported", func(t *testing.T) {
		content := `
mappings:
  - id: "parent.action"
    children:
      - "child.action"
    vscode:
      notSupported: true
  - id: "child.action"
    vscode:
      command: "child.cmd"
`
		reader := strings.NewReader(content)
		mc, err := load(reader)
		require.NoError(t, err)

		action, fallback := mc.GetExportAction("parent.action", pluginapi.EditorTypeVSCode)
		require.NotNil(t, action)
		assert.Equal(t, "child.action", action.ID)
		assert.True(t, fallback)
	})

	t.Run("falls back to first supported child in order", func(t *testing.T) {
		content := `
mappings:
  - id: "parent.action"
    children:
      - "child1.action"
      - "child2.action"
    vscode:
      notSupported: true
  - id: "child1.action"
    vscode:
      notSupported: true
  - id: "child2.action"
    vscode:
      command: "child2.cmd"
`
		reader := strings.NewReader(content)
		mc, err := load(reader)
		require.NoError(t, err)

		action, fallback := mc.GetExportAction("parent.action", pluginapi.EditorTypeVSCode)
		require.NotNil(t, action)
		assert.Equal(t, "child2.action", action.ID)
		assert.True(t, fallback)
	})

	t.Run("returns nil when no action supported", func(t *testing.T) {
		content := `
mappings:
  - id: "parent.action"
    children:
      - "child.action"
    vscode:
      notSupported: true
  - id: "child.action"
    vscode:
      notSupported: true
`
		reader := strings.NewReader(content)
		mc, err := load(reader)
		require.NoError(t, err)

		action, fallback := mc.GetExportAction("parent.action", pluginapi.EditorTypeVSCode)
		assert.Nil(t, action)
		assert.False(t, fallback)
	})

	t.Run("returns nil for non-existent action", func(t *testing.T) {
		content := `
mappings:
  - id: "some.action"
    vscode:
      command: "cmd"
`
		reader := strings.NewReader(content)
		mc, err := load(reader)
		require.NoError(t, err)

		action, fallback := mc.GetExportAction("non.existent", pluginapi.EditorTypeVSCode)
		assert.Nil(t, action)
		assert.False(t, fallback)
	})
}
