package helix

// helixConfig is a strong-typed representation of the subset of Helix's
// config.toml that we care about for keybindings.
// Only the [keys.*] sections are modeled here to keep scope focused.
// Other unrelated Helix configuration sections will be ignored on decode
// and are not emitted on encode.
//
// We keep modes dynamic to match Helix's design: [keys.<mode>]
// Example:
// [keys.normal]
//
//	"C-k" = "command"
//
// [keys.insert]
//
//	"M-c" = "another_command"
//
// See also: export.go/import.go for usage.
type helixConfig struct {
	Keys helixKeys `toml:"keys,omitempty"`
}

// HelixMode enumerates known Helix modes. Underlying type is string to match TOML keys.
// Unknown/custom modes are still representable as HelixMode values.
type HelixMode string

const (
	HelixModeNormal HelixMode = "normal"
	HelixModeInsert HelixMode = "insert"
	HelixModeSelect HelixMode = "select"
)

// helixKeys represents [keys.<mode>] => { "<key>": "<command>" }
// e.g. keys[HelixModeNormal]["M-c"] = "yank"
//
// We model known modes as struct fields for strong typing.

type helixKeys struct {
	Normal map[string]string `toml:"normal,omitempty"`
	Insert map[string]string `toml:"insert,omitempty"`
	Select map[string]string `toml:"select,omitempty"`
}
