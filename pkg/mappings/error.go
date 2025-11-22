package mappings

import (
	"encoding/json"
	"fmt"
)

type DuplicateActionMappingError struct {
	Editor string
	// editor action name -> all involved universal action ids
	Duplicates map[string][]string
}

func (e *DuplicateActionMappingError) Error() string {
	payload := map[string]any{
		"editor":     e.Editor,
		"duplicates": e.Duplicates,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("duplicate mapping error for %s: %v", e.Editor, err)
	}
	return string(b)
}
