package keycode

import (
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/bimap"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var keyCodeMap = bimap.NewBiMapFromMap(map[string]keymapv1.KeyCode{
	"a":              keymapv1.KeyCode_A,
	"b":              keymapv1.KeyCode_B,
	"c":              keymapv1.KeyCode_C,
	"d":              keymapv1.KeyCode_D,
	"e":              keymapv1.KeyCode_E,
	"f":              keymapv1.KeyCode_F,
	"g":              keymapv1.KeyCode_G,
	"h":              keymapv1.KeyCode_H,
	"i":              keymapv1.KeyCode_I,
	"j":              keymapv1.KeyCode_J,
	"k":              keymapv1.KeyCode_K,
	"l":              keymapv1.KeyCode_L,
	"m":              keymapv1.KeyCode_M,
	"n":              keymapv1.KeyCode_N,
	"o":              keymapv1.KeyCode_O,
	"p":              keymapv1.KeyCode_P,
	"q":              keymapv1.KeyCode_Q,
	"r":              keymapv1.KeyCode_R,
	"s":              keymapv1.KeyCode_S,
	"t":              keymapv1.KeyCode_T,
	"u":              keymapv1.KeyCode_U,
	"v":              keymapv1.KeyCode_V,
	"w":              keymapv1.KeyCode_W,
	"x":              keymapv1.KeyCode_X,
	"y":              keymapv1.KeyCode_Y,
	"z":              keymapv1.KeyCode_Z,
	"0":              keymapv1.KeyCode_DIGIT_0,
	"1":              keymapv1.KeyCode_DIGIT_1,
	"2":              keymapv1.KeyCode_DIGIT_2,
	"3":              keymapv1.KeyCode_DIGIT_3,
	"4":              keymapv1.KeyCode_DIGIT_4,
	"5":              keymapv1.KeyCode_DIGIT_5,
	"6":              keymapv1.KeyCode_DIGIT_6,
	"7":              keymapv1.KeyCode_DIGIT_7,
	"8":              keymapv1.KeyCode_DIGIT_8,
	"9":              keymapv1.KeyCode_DIGIT_9,
	"capslock":       keymapv1.KeyCode_CAPS_LOCK,
	"shift":          keymapv1.KeyCode_SHIFT,
	"fn":             keymapv1.KeyCode_FUNCTION,
	"ctrl":           keymapv1.KeyCode_CONTROL,
	"alt":            keymapv1.KeyCode_OPTION,
	"cmd":            keymapv1.KeyCode_COMMAND,
	"rightcmd":       keymapv1.KeyCode_RIGHT_COMMAND,
	"rightalt":       keymapv1.KeyCode_RIGHT_OPTION,
	"rightctrl":      keymapv1.KeyCode_RIGHT_CONTROL,
	"rightshift":     keymapv1.KeyCode_RIGHT_SHIFT,
	"enter":          keymapv1.KeyCode_RETURN,
	"\\":             keymapv1.KeyCode_BACKSLASH,
	"`":              keymapv1.KeyCode_BACKTICK,
	",":              keymapv1.KeyCode_COMMA,
	"=":              keymapv1.KeyCode_EQUAL,
	"-":              keymapv1.KeyCode_MINUS,
	"+":              keymapv1.KeyCode_PLUS,
	".":              keymapv1.KeyCode_PERIOD,
	"'":              keymapv1.KeyCode_QUOTE,
	";":              keymapv1.KeyCode_SEMICOLON,
	"/":              keymapv1.KeyCode_SLASH,
	"space":          keymapv1.KeyCode_SPACE,
	"tab":            keymapv1.KeyCode_TAB,
	"[":              keymapv1.KeyCode_LEFT_BRACKET,
	"]":              keymapv1.KeyCode_RIGHT_BRACKET,
	"pageup":         keymapv1.KeyCode_PAGE_UP,
	"pagedown":       keymapv1.KeyCode_PAGE_DOWN,
	"home":           keymapv1.KeyCode_HOME,
	"end":            keymapv1.KeyCode_END,
	"up":             keymapv1.KeyCode_UP_ARROW,
	"right":          keymapv1.KeyCode_RIGHT_ARROW,
	"down":           keymapv1.KeyCode_DOWN_ARROW,
	"left":           keymapv1.KeyCode_LEFT_ARROW,
	"escape":         keymapv1.KeyCode_ESCAPE,
	"backspace":      keymapv1.KeyCode_DELETE,
	"delete":         keymapv1.KeyCode_FORWARD_DELETE,
	"insert":         keymapv1.KeyCode_INSERT,
	"help":           keymapv1.KeyCode_HELP,
	"mute":           keymapv1.KeyCode_MUTE,
	"volumeup":       keymapv1.KeyCode_VOLUME_UP,
	"volumedown":     keymapv1.KeyCode_VOLUME_DOWN,
	"f1":             keymapv1.KeyCode_F1,
	"f2":             keymapv1.KeyCode_F2,
	"f3":             keymapv1.KeyCode_F3,
	"f4":             keymapv1.KeyCode_F4,
	"f5":             keymapv1.KeyCode_F5,
	"f6":             keymapv1.KeyCode_F6,
	"f7":             keymapv1.KeyCode_F7,
	"f8":             keymapv1.KeyCode_F8,
	"f9":             keymapv1.KeyCode_F9,
	"f10":            keymapv1.KeyCode_F10,
	"f11":            keymapv1.KeyCode_F11,
	"f12":            keymapv1.KeyCode_F12,
	"f13":            keymapv1.KeyCode_F13,
	"f14":            keymapv1.KeyCode_F14,
	"f15":            keymapv1.KeyCode_F15,
	"f16":            keymapv1.KeyCode_F16,
	"f17":            keymapv1.KeyCode_F17,
	"f18":            keymapv1.KeyCode_F18,
	"f19":            keymapv1.KeyCode_F19,
	"f20":            keymapv1.KeyCode_F20,
	"numpad0":        keymapv1.KeyCode_NUMPAD_0,
	"numpad1":        keymapv1.KeyCode_NUMPAD_1,
	"numpad2":        keymapv1.KeyCode_NUMPAD_2,
	"numpad3":        keymapv1.KeyCode_NUMPAD_3,
	"numpad4":        keymapv1.KeyCode_NUMPAD_4,
	"numpad5":        keymapv1.KeyCode_NUMPAD_5,
	"numpad6":        keymapv1.KeyCode_NUMPAD_6,
	"numpad7":        keymapv1.KeyCode_NUMPAD_7,
	"numpad8":        keymapv1.KeyCode_NUMPAD_8,
	"numpad9":        keymapv1.KeyCode_NUMPAD_9,
	"numpad_clear":   keymapv1.KeyCode_NUMPAD_CLEAR,
	"numpad_decimal": keymapv1.KeyCode_NUMPAD_DECIMAL,
	"numpad_divide":  keymapv1.KeyCode_NUMPAD_DIVIDE,
	"numpad_enter":   keymapv1.KeyCode_NUMPAD_ENTER,
	"numpad_equals":  keymapv1.KeyCode_NUMPAD_EQUALS,
	// as we are using vscode-like keymap, we use numpad_subtract instead of numpad_minus
	"numpad_subtract": keymapv1.KeyCode_NUMPAD_MINUS,
	"numpad_multiply": keymapv1.KeyCode_NUMPAD_MULTIPLY,
	"numpad_add":      keymapv1.KeyCode_NUMPAD_PLUS,
})

func FromString(s string) (keymapv1.KeyCode, bool) {
	return keyCodeMap.Get(strings.ToLower(s))
}

func ToString(kc keymapv1.KeyCode) (string, bool) {
	return keyCodeMap.GetInverse(kc)
}

func MustKeyCode(s string) keymapv1.KeyCode {
	kc, ok := FromString(s)
	if !ok {
		panic("keycode not found: " + s)
	}
	return kc
}
