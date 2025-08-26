package mappings

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func containsAll(slice []string, want ...string) bool {
	m := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		m[s] = struct{}{}
	}
	for _, w := range want {
		if _, ok := m[w]; !ok {
			return false
		}
	}
	return true
}

// -------------------- IntelliJ --------------------
func TestCheckIntelliJDuplicateConfig_NoDuplicates(t *testing.T) {
	mappings := map[string]ActionMappingConfig{
		"a": {IntelliJ: IntelliJMappingConfig{Action: "FocusEditor"}},
		"b": {IntelliJ: IntelliJMappingConfig{Action: "ActivateCommitToolWindow"}},
		"c": {IntelliJ: IntelliJMappingConfig{Action: ""}}, // ignored
	}
	assert.NoError(t, checkIntellijDuplicateConfig(mappings))
}

func TestCheckIntelliJDuplicateConfig_Duplicates(t *testing.T) {
	mappings := map[string]ActionMappingConfig{
		"x": {IntelliJ: IntelliJMappingConfig{Action: "FocusEditor"}},
		"y": {IntelliJ: IntelliJMappingConfig{Action: "FocusEditor"}},
		"z": {IntelliJ: IntelliJMappingConfig{Action: "Other"}},
	}
	err := checkIntellijDuplicateConfig(mappings)
	assert.Error(t, err)
	var derr *DuplicateActionMappingError
	assert.True(t, errors.As(err, &derr), fmt.Sprintf("expected DuplicateActionMappingError, got %T: %v", err, err))
	assert.Equal(t, "intellij", derr.Editor)
	got, ok := derr.Duplicates["FocusEditor"]
	assert.True(t, ok, fmt.Sprintf("expected duplicates for FocusEditor, got keys: %v", keys(derr.Duplicates)))
	assert.True(t, len(got) >= 2, fmt.Sprintf("expected at least 2 ids, got %v", got))
	assert.True(t, containsAll(got, "x", "y"), fmt.Sprintf("expected ids [x y] in any order, got %v", got))
}

// -------------------- VSCode --------------------
func TestCheckVscodeDuplicateConfig_NoDuplicates(t *testing.T) {
	mappings := map[string]ActionMappingConfig{
		"a": {VSCode: VscodeConfigs{{Command: "editor.action.goToDefinition", When: "editorTextFocus"}}},
		"b": {VSCode: VscodeConfigs{{Command: "workbench.action.quickOpen", When: "inputFocus"}}},
	}
	assert.NoError(t, checkVscodeDuplicateConfig(mappings))
}

func TestCheckVscodeDuplicateConfig_Duplicates(t *testing.T) {
	args := map[string]interface{}{"a": 1, "b": "x"}
	mappings := map[string]ActionMappingConfig{
		"id1": {VSCode: VscodeConfigs{{Command: "workbench.action.some", When: "editorTextFocus", Args: args}}},
		"id2": {VSCode: VscodeConfigs{{Command: "workbench.action.some", When: "editorTextFocus", Args: args}}},
	}
	err := checkVscodeDuplicateConfig(mappings)
	assert.Error(t, err)
	var derr *DuplicateActionMappingError
	assert.True(t, errors.As(err, &derr), fmt.Sprintf("expected DuplicateActionMappingError, got %T: %v", err, err))
	assert.Equal(t, "vscode", derr.Editor)
	// Build expected dup key exactly as code does
	argsBytes, _ := json.Marshal(args)
	key := fmt.Sprintf(`{"command":%q,"when":%q,"args":%q}`, "workbench.action.some", "editorTextFocus", string(argsBytes))
	got, ok := derr.Duplicates[key]
	assert.True(t, ok, fmt.Sprintf("expected duplicates for key %s, got keys: %v", key, keys(derr.Duplicates)))
	assert.True(t, len(got) >= 2, fmt.Sprintf("expected at least 2 ids, got %v", got))
	assert.True(t, containsAll(got, "id1", "id2"), fmt.Sprintf("expected ids [id1 id2] in any order, got %v", got))
}

// -------------------- Zed --------------------
func TestCheckZedDuplicateConfig_NoDuplicates(t *testing.T) {
	mappings := map[string]ActionMappingConfig{
		"a": {Zed: ZedConfigs{{Action: "editor:go-to-definition", Context: "Editor"}}},
		"b": {Zed: ZedConfigs{{Action: "window:show-quick-open", Context: "Global"}}},
	}
	assert.NoError(t, checkZedDuplicateConfig(mappings))
}

func TestCheckZedDuplicateConfig_Duplicates(t *testing.T) {
	args := map[string]interface{}{"x": true}
	mappings := map[string]ActionMappingConfig{
		"k1": {Zed: ZedConfigs{{Action: "editor:go-to-definition", Context: "Editor", Args: args}}},
		"k2": {Zed: ZedConfigs{{Action: "editor:go-to-definition", Context: "Editor", Args: args}}},
	}
	err := checkZedDuplicateConfig(mappings)
	assert.Error(t, err)
	var derr *DuplicateActionMappingError
	assert.True(t, errors.As(err, &derr), fmt.Sprintf("expected DuplicateActionMappingError, got %T: %v", err, err))
	assert.Equal(t, "zed", derr.Editor)
	argsBytes, _ := json.Marshal(args)
	key := fmt.Sprintf(`{"action":%q,"context":%q,"args":%q}`, "editor:go-to-definition", "Editor", string(argsBytes))
	got, ok := derr.Duplicates[key]
	assert.True(t, ok, fmt.Sprintf("expected duplicates for key %s, got keys: %v", key, keys(derr.Duplicates)))
	assert.True(t, len(got) >= 2, fmt.Sprintf("expected at least 2 ids, got %v", got))
	assert.True(t, containsAll(got, "k1", "k2"), fmt.Sprintf("expected ids [k1 k2] in any order, got %v", got))
}

func keys(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
