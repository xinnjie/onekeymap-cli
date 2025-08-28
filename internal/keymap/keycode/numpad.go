package keycode

import keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"

// IsNumpad checks if a given key code is a numpad key.
func IsNumpad(kc keymapv1.KeyCode) bool {
	if kc < keymapv1.KeyCode_NUMPAD_0 {
		return false
	}
	if kc > keymapv1.KeyCode_NUMPAD_INSERT {
		return false
	}
	return true
}
