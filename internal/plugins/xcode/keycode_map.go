package xcode

import (
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// Xcode key code constants
const (
	// Escape key
	xcodeEscape = 0x1B

	// Arrow keys (Unicode private use area)
	xcodeUpArrow    = 0xF700
	xcodeDownArrow  = 0xF701
	xcodeLeftArrow  = 0xF702
	xcodeRightArrow = 0xF703

	// Function keys (Unicode private use area)
	xcodeF1  = 0xF704
	xcodeF2  = 0xF705
	xcodeF3  = 0xF706
	xcodeF4  = 0xF707
	xcodeF5  = 0xF708
	xcodeF6  = 0xF709
	xcodeF7  = 0xF70A
	xcodeF8  = 0xF70B
	xcodeF9  = 0xF70C
	xcodeF10 = 0xF70D
	xcodeF11 = 0xF70E
	xcodeF12 = 0xF70F
	xcodeF13 = 0xF710
	xcodeF14 = 0xF711
	xcodeF15 = 0xF712
	xcodeF16 = 0xF713
	xcodeF17 = 0xF714
	xcodeF18 = 0xF715
	xcodeF19 = 0xF716
	xcodeF20 = 0xF717
	xcodeF21 = 0xF727
	xcodeF22 = 0xF728
	xcodeF23 = 0xF729
	xcodeF24 = 0xF72B
	xcodeF25 = 0xF72C
	xcodeF26 = 0xF72D

	// Letter keys (hex values for A-Z)
	xcodeA = 0x41
	xcodeC = 0x43
	xcodeE = 0x45
	xcodeG = 0x47
	xcodeK = 0x4B
	xcodeL = 0x4C
	xcodeN = 0x4E
	xcodeQ = 0x51
	xcodeR = 0x52
	xcodeS = 0x53
	xcodeT = 0x54
	xcodeU = 0x55
	xcodeV = 0x56
	xcodeW = 0x57
	xcodeY = 0x58
	xcodeZ = 0x59

	// Special keys
	xcodeLeftBracket = 0x5B
	xcodeBackslash   = 0x5C
)

// getRuneFromKeyCode converts a KeyCode directly to its Xcode rune representation.
// Reference: NSEvent.h in AppKit framework, see http://xahlee.info/kbd/i/NSEvent.h
//
//nolint:gocyclo,cyclop
func getRuneFromKeyCode(kc keymapv1.KeyCode) (rune, bool) {
	switch kc {
	// Letters A-Z - convert to lowercase
	case keymapv1.KeyCode_A:
		return 'a', true
	case keymapv1.KeyCode_B:
		return 'b', true
	case keymapv1.KeyCode_C:
		return 'c', true
	case keymapv1.KeyCode_D:
		return 'd', true
	case keymapv1.KeyCode_E:
		return 'e', true
	case keymapv1.KeyCode_F:
		return 'f', true
	case keymapv1.KeyCode_G:
		return 'g', true
	case keymapv1.KeyCode_H:
		return 'h', true
	case keymapv1.KeyCode_I:
		return 'i', true
	case keymapv1.KeyCode_J:
		return 'j', true
	case keymapv1.KeyCode_K:
		return 'k', true
	case keymapv1.KeyCode_L:
		return 'l', true
	case keymapv1.KeyCode_M:
		return 'm', true
	case keymapv1.KeyCode_N:
		return 'n', true
	case keymapv1.KeyCode_O:
		return 'o', true
	case keymapv1.KeyCode_P:
		return 'p', true
	case keymapv1.KeyCode_Q:
		return 'q', true
	case keymapv1.KeyCode_R:
		return 'r', true
	case keymapv1.KeyCode_S:
		return 's', true
	case keymapv1.KeyCode_T:
		return 't', true
	case keymapv1.KeyCode_U:
		return 'u', true
	case keymapv1.KeyCode_V:
		return 'v', true
	case keymapv1.KeyCode_W:
		return 'w', true
	case keymapv1.KeyCode_X:
		return 'x', true
	case keymapv1.KeyCode_Y:
		return 'y', true
	case keymapv1.KeyCode_Z:
		return 'z', true

	// Numbers 0-9
	case keymapv1.KeyCode_DIGIT_0:
		return '0', true
	case keymapv1.KeyCode_DIGIT_1:
		return '1', true
	case keymapv1.KeyCode_DIGIT_2:
		return '2', true
	case keymapv1.KeyCode_DIGIT_3:
		return '3', true
	case keymapv1.KeyCode_DIGIT_4:
		return '4', true
	case keymapv1.KeyCode_DIGIT_5:
		return '5', true
	case keymapv1.KeyCode_DIGIT_6:
		return '6', true
	case keymapv1.KeyCode_DIGIT_7:
		return '7', true
	case keymapv1.KeyCode_DIGIT_8:
		return '8', true
	case keymapv1.KeyCode_DIGIT_9:
		return '9', true

	// Common ASCII control/whitespace keys
	case keymapv1.KeyCode_TAB:
		return '\t', true
	case keymapv1.KeyCode_RETURN:
		return '\r', true
	case keymapv1.KeyCode_DELETE: // Backspace
		return '\b', true
	case keymapv1.KeyCode_ESCAPE:
		return xcodeEscape, true
	case keymapv1.KeyCode_SPACE:
		return ' ', true

	// Arrow Keys (Unicode private use area)
	case keymapv1.KeyCode_UP_ARROW:
		return xcodeUpArrow, true
	case keymapv1.KeyCode_DOWN_ARROW:
		return xcodeDownArrow, true
	case keymapv1.KeyCode_LEFT_ARROW:
		return xcodeLeftArrow, true
	case keymapv1.KeyCode_RIGHT_ARROW:
		return xcodeRightArrow, true

	// Function Keys (F1-F20)
	case keymapv1.KeyCode_F1:
		return xcodeF1, true
	case keymapv1.KeyCode_F2:
		return xcodeF2, true
	case keymapv1.KeyCode_F3:
		return xcodeF3, true
	case keymapv1.KeyCode_F4:
		return xcodeF4, true
	case keymapv1.KeyCode_F5:
		return xcodeF5, true
	case keymapv1.KeyCode_F6:
		return xcodeF6, true
	case keymapv1.KeyCode_F7:
		return xcodeF7, true
	case keymapv1.KeyCode_F8:
		return xcodeF8, true
	case keymapv1.KeyCode_F9:
		return xcodeF9, true
	case keymapv1.KeyCode_F10:
		return xcodeF10, true
	case keymapv1.KeyCode_F11:
		return xcodeF11, true
	case keymapv1.KeyCode_F12:
		return xcodeF12, true
	case keymapv1.KeyCode_F13:
		return xcodeF13, true
	case keymapv1.KeyCode_F14:
		return xcodeF14, true
	case keymapv1.KeyCode_F15:
		return xcodeF15, true
	case keymapv1.KeyCode_F16:
		return xcodeF16, true
	case keymapv1.KeyCode_F17:
		return xcodeF17, true
	case keymapv1.KeyCode_F18:
		return xcodeF18, true
	case keymapv1.KeyCode_F19:
		return xcodeF19, true
	case keymapv1.KeyCode_F20:
		return xcodeF20, true

	// Navigation Keys
	case keymapv1.KeyCode_INSERT_OR_HELP:
		return xcodeF21, true
	case keymapv1.KeyCode_FORWARD_DELETE:
		return xcodeF22, true
	case keymapv1.KeyCode_HOME:
		return xcodeF23, true
	case keymapv1.KeyCode_END:
		return xcodeF24, true
	case keymapv1.KeyCode_PAGE_UP:
		return xcodeF25, true
	case keymapv1.KeyCode_PAGE_DOWN:
		return xcodeF26, true

	// Keypad (numpad) keys via HIToolbox virtual key codes
	case keymapv1.KeyCode_NUMPAD_DECIMAL:
		return xcodeA, true
	case keymapv1.KeyCode_NUMPAD_MULTIPLY:
		return xcodeC, true
	case keymapv1.KeyCode_NUMPAD_PLUS:
		return xcodeE, true
	case keymapv1.KeyCode_NUMPAD_CLEAR:
		return xcodeG, true
	case keymapv1.KeyCode_NUMPAD_DIVIDE:
		return xcodeK, true
	case keymapv1.KeyCode_NUMPAD_ENTER:
		return xcodeL, true
	case keymapv1.KeyCode_NUMPAD_MINUS:
		return xcodeN, true
	case keymapv1.KeyCode_NUMPAD_EQUALS:
		return xcodeQ, true
	case keymapv1.KeyCode_NUMPAD_0:
		return xcodeR, true
	case keymapv1.KeyCode_NUMPAD_1:
		return xcodeS, true
	case keymapv1.KeyCode_NUMPAD_2:
		return xcodeT, true
	case keymapv1.KeyCode_NUMPAD_3:
		return xcodeU, true
	case keymapv1.KeyCode_NUMPAD_4:
		return xcodeV, true
	case keymapv1.KeyCode_NUMPAD_5:
		return xcodeW, true
	case keymapv1.KeyCode_NUMPAD_6:
		return xcodeY, true
	case keymapv1.KeyCode_NUMPAD_7:
		return xcodeZ, true
	case keymapv1.KeyCode_NUMPAD_8:
		return xcodeLeftBracket, true
	case keymapv1.KeyCode_NUMPAD_9:
		return xcodeBackslash, true

	// Unsupported KeyCodes by Xcode keybindings
	case keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
		keymapv1.KeyCode_CAPS_LOCK,
		keymapv1.KeyCode_SHIFT,
		keymapv1.KeyCode_FUNCTION,
		keymapv1.KeyCode_CONTROL,
		keymapv1.KeyCode_OPTION,
		keymapv1.KeyCode_COMMAND,
		keymapv1.KeyCode_RIGHT_COMMAND,
		keymapv1.KeyCode_RIGHT_OPTION,
		keymapv1.KeyCode_RIGHT_CONTROL,
		keymapv1.KeyCode_RIGHT_SHIFT,
		keymapv1.KeyCode_BACKSLASH,
		keymapv1.KeyCode_BACKTICK,
		keymapv1.KeyCode_COMMA,
		keymapv1.KeyCode_EQUAL,
		keymapv1.KeyCode_MINUS,
		keymapv1.KeyCode_PLUS,
		keymapv1.KeyCode_PERIOD,
		keymapv1.KeyCode_QUOTE,
		keymapv1.KeyCode_SEMICOLON,
		keymapv1.KeyCode_SLASH,
		keymapv1.KeyCode_LEFT_BRACKET,
		keymapv1.KeyCode_RIGHT_BRACKET,
		keymapv1.KeyCode_MUTE,
		keymapv1.KeyCode_VOLUME_UP,
		keymapv1.KeyCode_VOLUME_DOWN,
		keymapv1.KeyCode_NUMPAD_INSERT:
		return 0, false

	default:
		// This should never happen if all KeyCode enum values are covered
		return 0, false
	}
}

// getKeyCodeFromRune converts an Xcode rune to its KeyCode representation.
// This is the reverse operation of getRuneFromKeyCode.
// Note: This function only handles special keys. Normal letters and digits
// are handled directly in ParseKeybinding to avoid conflicts with numpad keys.
//
//nolint:gocyclo,cyclop
func getKeyCodeFromRune(r rune) (keymapv1.KeyCode, bool) {
	switch r {
	// Common ASCII control/whitespace keys
	case '\t':
		return keymapv1.KeyCode_TAB, true
	case '\r':
		return keymapv1.KeyCode_RETURN, true
	case '\b':
		return keymapv1.KeyCode_DELETE, true
	case xcodeEscape:
		return keymapv1.KeyCode_ESCAPE, true
	case ' ':
		return keymapv1.KeyCode_SPACE, true

	// Arrow Keys (Unicode private use area)
	case xcodeUpArrow:
		return keymapv1.KeyCode_UP_ARROW, true
	case xcodeDownArrow:
		return keymapv1.KeyCode_DOWN_ARROW, true
	case xcodeLeftArrow:
		return keymapv1.KeyCode_LEFT_ARROW, true
	case xcodeRightArrow:
		return keymapv1.KeyCode_RIGHT_ARROW, true

	// Function Keys (Unicode private use area)
	case xcodeF1:
		return keymapv1.KeyCode_F1, true
	case xcodeF2:
		return keymapv1.KeyCode_F2, true
	case xcodeF3:
		return keymapv1.KeyCode_F3, true
	case xcodeF4:
		return keymapv1.KeyCode_F4, true
	case xcodeF5:
		return keymapv1.KeyCode_F5, true
	case xcodeF6:
		return keymapv1.KeyCode_F6, true
	case xcodeF7:
		return keymapv1.KeyCode_F7, true
	case xcodeF8:
		return keymapv1.KeyCode_F8, true
	case xcodeF9:
		return keymapv1.KeyCode_F9, true
	case xcodeF10:
		return keymapv1.KeyCode_F10, true
	case xcodeF11:
		return keymapv1.KeyCode_F11, true
	case xcodeF12:
		return keymapv1.KeyCode_F12, true
	case xcodeF13:
		return keymapv1.KeyCode_F13, true
	case xcodeF14:
		return keymapv1.KeyCode_F14, true
	case xcodeF15:
		return keymapv1.KeyCode_F15, true
	case xcodeF16:
		return keymapv1.KeyCode_F16, true
	case xcodeF17:
		return keymapv1.KeyCode_F17, true
	case xcodeF18:
		return keymapv1.KeyCode_F18, true
	case xcodeF19:
		return keymapv1.KeyCode_F19, true
	case xcodeF20:
		return keymapv1.KeyCode_F20, true

	// Navigation Keys (Unicode private use area)
	case xcodeF21:
		return keymapv1.KeyCode_INSERT_OR_HELP, true
	case xcodeF22:
		return keymapv1.KeyCode_FORWARD_DELETE, true
	case xcodeF23:
		return keymapv1.KeyCode_HOME, true
	case xcodeF24:
		return keymapv1.KeyCode_END, true
	case xcodeF25:
		return keymapv1.KeyCode_PAGE_UP, true
	case xcodeF26:
		return keymapv1.KeyCode_PAGE_DOWN, true

	// Keypad keys (HIToolbox virtual key codes)
	// Note: These conflict with ASCII letters, but are distinguished by context
	case xcodeA:
		return keymapv1.KeyCode_NUMPAD_DECIMAL, true
	case xcodeC:
		return keymapv1.KeyCode_NUMPAD_MULTIPLY, true
	case xcodeE:
		return keymapv1.KeyCode_NUMPAD_PLUS, true
	case xcodeG:
		return keymapv1.KeyCode_NUMPAD_CLEAR, true
	case xcodeK:
		return keymapv1.KeyCode_NUMPAD_DIVIDE, true
	case xcodeL:
		return keymapv1.KeyCode_NUMPAD_ENTER, true
	case xcodeN:
		return keymapv1.KeyCode_NUMPAD_MINUS, true
	case xcodeQ:
		return keymapv1.KeyCode_NUMPAD_EQUALS, true
	case xcodeR:
		return keymapv1.KeyCode_NUMPAD_0, true
	case xcodeS:
		return keymapv1.KeyCode_NUMPAD_1, true
	case xcodeT:
		return keymapv1.KeyCode_NUMPAD_2, true
	case xcodeU:
		return keymapv1.KeyCode_NUMPAD_3, true
	case xcodeV:
		return keymapv1.KeyCode_NUMPAD_4, true
	case xcodeW:
		return keymapv1.KeyCode_NUMPAD_5, true
	case xcodeY:
		return keymapv1.KeyCode_NUMPAD_6, true
	case xcodeZ:
		return keymapv1.KeyCode_NUMPAD_7, true
	case xcodeLeftBracket:
		return keymapv1.KeyCode_NUMPAD_8, true
	case xcodeBackslash:
		return keymapv1.KeyCode_NUMPAD_9, true

	default:
		return keymapv1.KeyCode_KEY_CODE_UNSPECIFIED, false
	}
}
