package internal

import (
	"github.com/tailscale/hujson"
)

// StripJSONComments converts JSONC (with //, /* */ comments and trailing commas)
// into standard JSON to make it safe for encoding/json.
func StripJSONComments(data []byte) []byte {
	// hujson.Standardize handles comments and trailing commas robustly.
	std, err := hujson.Standardize(data)
	if err != nil {
		// Fallback to original content if we cannot parse; caller will surface error.
		return data
	}
	return std
}
