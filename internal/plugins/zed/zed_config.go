package zed

import "encoding/json"

type zedKeymapConfig = []zedKeymapOfContext

// zedKeymapOfContext represents the structure of Zed's keymap.json file.
type zedKeymapOfContext struct {
	Context  string                    `json:"context,omitempty"`
	Bindings map[string]zedActionValue `json:"bindings"`
}

// zedActionValue is a strong-typed action representation for export.
// It marshals to JSON as either:
// - a string: "action"
// - or an array: ["action", {args}]
// e.g.
// "ctrl-shift-r": [
//
//	"pane::DeploySearch",
//	{
//	  "replace_enabled": true
//	}
//
// ],
// action: "pane::DeploySearch"
// args: {"replace_enabled": true}
type zedActionValue struct {
	Action string
	Args   map[string]interface{}
}

func (z zedActionValue) MarshalJSON() ([]byte, error) {
	if len(z.Args) > 0 {
		return json.Marshal([]interface{}{z.Action, z.Args})
	}
	return json.Marshal(z.Action)
}

// UnmarshalJSON accepts either a string ("action") or an array format
// ["action", {args}] and maps it into the strong type. Invalid shapes are
// treated as zero value without returning an error to allow robust import.
func (z *zedActionValue) UnmarshalJSON(data []byte) error {
	// Try simple string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		z.Action = s
		z.Args = nil
		return nil
	}

	// Try array format: [action, args]
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) == 0 {
			// empty array => ignore

			return nil
		}
		var act string
		if err := json.Unmarshal(arr[0], &act); err != nil {
			// invalid first element => ignore
			//nolint:nilerr // tolerate malformed entry; treat as zero value to keep importer resilient
			return nil
		}
		z.Action = act
		if len(arr) > 1 {
			var args map[string]interface{}
			if err := json.Unmarshal(arr[1], &args); err == nil {
				z.Args = args
			}
		}
		return nil
	}

	// Unsupported JSON shape => ignore
	return nil
}
