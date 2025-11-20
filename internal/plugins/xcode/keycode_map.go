package xcode

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
)

// Xcode .idekeybindings config key code constants
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
)

// getRuneFromKeyCode converts a KeyCode directly to its Xcode rune representation.
// Reference: NSEvent.h in AppKit framework, see http://xahlee.info/kbd/i/NSEvent.h
//
//nolint:gocyclo,cyclop
func getRuneFromKeyCode(kc keycode.KeyCode) (rune, bool) {
	switch kc {
	// Letters A-Z - convert to lowercase
	case keycode.KeyCodeA:
		return 'a', true
	case keycode.KeyCodeB:
		return 'b', true
	case keycode.KeyCodeC:
		return 'c', true
	case keycode.KeyCodeD:
		return 'd', true
	case keycode.KeyCodeE:
		return 'e', true
	case keycode.KeyCodeF:
		return 'f', true
	case keycode.KeyCodeG:
		return 'g', true
	case keycode.KeyCodeH:
		return 'h', true
	case keycode.KeyCodeI:
		return 'i', true
	case keycode.KeyCodeJ:
		return 'j', true
	case keycode.KeyCodeK:
		return 'k', true
	case keycode.KeyCodeL:
		return 'l', true
	case keycode.KeyCodeM:
		return 'm', true
	case keycode.KeyCodeN:
		return 'n', true
	case keycode.KeyCodeO:
		return 'o', true
	case keycode.KeyCodeP:
		return 'p', true
	case keycode.KeyCodeQ:
		return 'q', true
	case keycode.KeyCodeR:
		return 'r', true
	case keycode.KeyCodeS:
		return 's', true
	case keycode.KeyCodeT:
		return 't', true
	case keycode.KeyCodeU:
		return 'u', true
	case keycode.KeyCodeV:
		return 'v', true
	case keycode.KeyCodeW:
		return 'w', true
	case keycode.KeyCodeX:
		return 'x', true
	case keycode.KeyCodeY:
		return 'y', true
	case keycode.KeyCodeZ:
		return 'z', true

	// Numbers 0-9
	case keycode.KeyCodeDigit0:
		return '0', true
	case keycode.KeyCodeDigit1:
		return '1', true
	case keycode.KeyCodeDigit2:
		return '2', true
	case keycode.KeyCodeDigit3:
		return '3', true
	case keycode.KeyCodeDigit4:
		return '4', true
	case keycode.KeyCodeDigit5:
		return '5', true
	case keycode.KeyCodeDigit6:
		return '6', true
	case keycode.KeyCodeDigit7:
		return '7', true
	case keycode.KeyCodeDigit8:
		return '8', true
	case keycode.KeyCodeDigit9:
		return '9', true

	// Common ASCII control/whitespace keys
	case keycode.KeyCodeTab:
		return '\t', true
	case keycode.KeyCodeEnter:
		return '\r', true
	case keycode.KeyCodeBackspace:
		return '\b', true
	case keycode.KeyCodeEscape:
		return xcodeEscape, true
	case keycode.KeyCodeSpace:
		return ' ', true

	// Arrow Keys (Unicode private use area)
	case keycode.KeyCodeUp:
		return xcodeUpArrow, true
	case keycode.KeyCodeDown:
		return xcodeDownArrow, true
	case keycode.KeyCodeLeft:
		return xcodeLeftArrow, true
	case keycode.KeyCodeRight:
		return xcodeRightArrow, true

	// Function Keys (F1-F20)
	case keycode.KeyCodeF1:
		return xcodeF1, true
	case keycode.KeyCodeF2:
		return xcodeF2, true
	case keycode.KeyCodeF3:
		return xcodeF3, true
	case keycode.KeyCodeF4:
		return xcodeF4, true
	case keycode.KeyCodeF5:
		return xcodeF5, true
	case keycode.KeyCodeF6:
		return xcodeF6, true
	case keycode.KeyCodeF7:
		return xcodeF7, true
	case keycode.KeyCodeF8:
		return xcodeF8, true
	case keycode.KeyCodeF9:
		return xcodeF9, true
	case keycode.KeyCodeF10:
		return xcodeF10, true
	case keycode.KeyCodeF11:
		return xcodeF11, true
	case keycode.KeyCodeF12:
		return xcodeF12, true
	case keycode.KeyCodeF13:
		return xcodeF13, true
	case keycode.KeyCodeF14:
		return xcodeF14, true
	case keycode.KeyCodeF15:
		return xcodeF15, true
	case keycode.KeyCodeF16:
		return xcodeF16, true
	case keycode.KeyCodeF17:
		return xcodeF17, true
	case keycode.KeyCodeF18:
		return xcodeF18, true
	case keycode.KeyCodeF19:
		return xcodeF19, true
	case keycode.KeyCodeF20:
		return xcodeF20, true

	// Navigation Keys
	case keycode.KeyCodeInsert:
		return xcodeF21, true
	case keycode.KeyCodeDelete:
		return xcodeF22, true
	case keycode.KeyCodeHome:
		return xcodeF23, true
	case keycode.KeyCodeEnd:
		return xcodeF24, true
	case keycode.KeyCodePageUp:
		return xcodeF25, true
	case keycode.KeyCodePageDown:
		return xcodeF26, true

	// FIXME(xinnjie): Do not have keyboard with numpad. So can not test it.
	case keycode.KeyCodeNumpadDecimal:
		return '.', true
	case keycode.KeyCodeNumpadMultiply:
		return '*', true
	case keycode.KeyCodeNumpadAdd:
		return '+', true
	case keycode.KeyCodeNumpadClear:
		return 0, false
	case keycode.KeyCodeNumpadDivide:
		return '/', true
	case keycode.KeyCodeNumpadEnter:
		return '\r', true
	case keycode.KeyCodeNumpadSubtract:
		return '-', true
	case keycode.KeyCodeNumpadEquals:
		return '=', true
	case keycode.KeyCodeNumpad0:
		return '0', true
	case keycode.KeyCodeNumpad1:
		return '1', true
	case keycode.KeyCodeNumpad2:
		return '2', true
	case keycode.KeyCodeNumpad3:
		return '3', true
	case keycode.KeyCodeNumpad4:
		return '4', true
	case keycode.KeyCodeNumpad5:
		return '5', true
	case keycode.KeyCodeNumpad6:
		return '6', true
	case keycode.KeyCodeNumpad7:
		return '7', true
	case keycode.KeyCodeNumpad8:
		return '8', true
	case keycode.KeyCodeNumpad9:
		return '9', true

	case keycode.KeyCodeBackslash:
		return '\\', true
	case keycode.KeyCodeBacktick:
		return '`', true
	case keycode.KeyCodeComma:
		return ',', true
	case keycode.KeyCodeEqual:
		return '=', true
	case keycode.KeyCodeMinus:
		return '-', true
	case keycode.KeyCodePeriod:
		return '.', true
	case keycode.KeyCodeQuote:
		return '\'', true
	case keycode.KeyCodeSemicolon:
		return ';', true
	case keycode.KeyCodeSlash:
		return '/', true
	case keycode.KeyCodeLeftBracket:
		return '[', true
	case keycode.KeyCodeRightBracket:
		return ']', true

	// Unsupported KeyCodes by Xcode keybindings
	case keycode.KeyCodeCapsLock,
		keycode.KeyCodeShift,
		keycode.KeyCodeFn,
		keycode.KeyCodeCtrl,
		keycode.KeyCodeAlt,
		keycode.KeyCodeCmd,
		keycode.KeyCodeRightCmd,
		keycode.KeyCodeRightAlt,
		keycode.KeyCodeRightCtrl,
		keycode.KeyCodeRightShift,
		keycode.KeyCodeMute,
		keycode.KeyCodeVolumeUp,
		keycode.KeyCodeVolumeDown:
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
func getKeyCodeFromRune(r rune) (keycode.KeyCode, bool) {
	switch r {
	// Common ASCII control/whitespace keys
	case '\t':
		return keycode.KeyCodeTab, true
	case '\r':
		return keycode.KeyCodeEnter, true
	case '\b':
		return keycode.KeyCodeDelete, true
	case xcodeEscape:
		return keycode.KeyCodeEscape, true
	case ' ':
		return keycode.KeyCodeSpace, true

	// Arrow Keys (Unicode private use area)
	case xcodeUpArrow:
		return keycode.KeyCodeUp, true
	case xcodeDownArrow:
		return keycode.KeyCodeDown, true
	case xcodeLeftArrow:
		return keycode.KeyCodeLeft, true
	case xcodeRightArrow:
		return keycode.KeyCodeRight, true

	// Function Keys (Unicode private use area)
	case xcodeF1:
		return keycode.KeyCodeF1, true
	case xcodeF2:
		return keycode.KeyCodeF2, true
	case xcodeF3:
		return keycode.KeyCodeF3, true
	case xcodeF4:
		return keycode.KeyCodeF4, true
	case xcodeF5:
		return keycode.KeyCodeF5, true
	case xcodeF6:
		return keycode.KeyCodeF6, true
	case xcodeF7:
		return keycode.KeyCodeF7, true
	case xcodeF8:
		return keycode.KeyCodeF8, true
	case xcodeF9:
		return keycode.KeyCodeF9, true
	case xcodeF10:
		return keycode.KeyCodeF10, true
	case xcodeF11:
		return keycode.KeyCodeF11, true
	case xcodeF12:
		return keycode.KeyCodeF12, true
	case xcodeF13:
		return keycode.KeyCodeF13, true
	case xcodeF14:
		return keycode.KeyCodeF14, true
	case xcodeF15:
		return keycode.KeyCodeF15, true
	case xcodeF16:
		return keycode.KeyCodeF16, true
	case xcodeF17:
		return keycode.KeyCodeF17, true
	case xcodeF18:
		return keycode.KeyCodeF18, true
	case xcodeF19:
		return keycode.KeyCodeF19, true
	case xcodeF20:
		return keycode.KeyCodeF20, true

	// Navigation Keys (Unicode private use area)
	case xcodeF21:
		return keycode.KeyCodeInsert, true
	case xcodeF22:
		return keycode.KeyCodeDelete, true
	case xcodeF23:
		return keycode.KeyCodeHome, true
	case xcodeF24:
		return keycode.KeyCodeEnd, true
	case xcodeF25:
		return keycode.KeyCodePageUp, true
	case xcodeF26:
		return keycode.KeyCodePageDown, true

	default:
		return "", false
	}
}
