package mappings

type IntelliJMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	Action string `yaml:"action"`
}

func checkIntellijDuplicateConfig(mappings map[string]ActionMappingConfig) error {
	seen := make(map[string]string)   // action -> first seen universal action id
	dups := make(map[string][]string) // action -> all involved universal action ids (including first)
	for id, m := range mappings {
		action := m.IntelliJ.Action
		if action == "" {
			continue
		}
		// Skip configs that are disabled for import (export-only)
		if m.IntelliJ.DisableImport {
			continue
		}
		if prev, ok := seen[action]; ok {
			if _, exists := dups[action]; !exists {
				dups[action] = []string{prev}
			}
			dups[action] = append(dups[action], id)
			continue
		}
		seen[action] = id
	}
	if len(dups) == 0 {
		return nil
	}
	return &DuplicateActionMappingError{Editor: "intellij", Duplicates: dups}
}
