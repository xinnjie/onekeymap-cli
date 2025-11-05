package xcode

// xcodeKeyCodeMap returns a map of Unicode private use area characters to key names.
// These are the special function keys used in Xcode .idekeybindings files
// Reference: NSEvent.h in AppKit framework, see http://xahlee.info/kbd/i/NSEvent.h
func xcodeKeyCodeMap() map[rune]string {
	return map[rune]string{
		// Arrow Keys
		0xF700: "up",
		0xF701: "down",
		0xF702: "left",
		0xF703: "right",

		// Function Keys F1-F35
		0xF704: "f1",
		0xF705: "f2",
		0xF706: "f3",
		0xF707: "f4",
		0xF708: "f5",
		0xF709: "f6",
		0xF70A: "f7",
		0xF70B: "f8",
		0xF70C: "f9",
		0xF70D: "f10",
		0xF70E: "f11",
		0xF70F: "f12",
		0xF710: "f13",
		0xF711: "f14",
		0xF712: "f15",
		0xF713: "f16",
		0xF714: "f17",
		0xF715: "f18",
		0xF716: "f19",
		0xF717: "f20",
		0xF718: "f21",
		0xF719: "f22",
		0xF71A: "f23",
		0xF71B: "f24",
		0xF71C: "f25",
		0xF71D: "f26",
		0xF71E: "f27",
		0xF71F: "f28",
		0xF720: "f29",
		0xF721: "f30",
		0xF722: "f31",
		0xF723: "f32",
		0xF724: "f33",
		0xF725: "f34",
		0xF726: "f35",

		// Navigation Keys
		0xF727: "insert",
		0xF728: "delete", // Forward delete (Fn+Delete)
		0xF729: "home",
		0xF72A: "begin",
		0xF72B: "end",
		0xF72C: "pageup",
		0xF72D: "pagedown",

		// Other Special Keys
		0xF72E: "printscreen",
		0xF72F: "scrolllock",
		0xF730: "pause",
		0xF731: "sysreq",
		0xF732: "break",
		0xF733: "reset",
		0xF734: "stop",
		0xF735: "menu",
		0xF736: "user",
		0xF737: "system",
		0xF738: "print",
		0xF739: "clearline",
		0xF73A: "cleardisplay",
		0xF73B: "insertline",
		0xF73C: "deleteline",
		0xF73D: "insertchar",
		0xF73E: "deletechar",
		0xF73F: "prev",
		0xF740: "next",
		0xF741: "select",
		0xF742: "execute",
		0xF743: "undo",
		0xF744: "redo",
		0xF745: "find",
		0xF746: "help",
		0xF747: "modeswitch",
	}
}

// getKeyNameFromCode returns the key name for a Unicode code point
func getKeyNameFromCode(r rune) (string, bool) {
	name, ok := xcodeKeyCodeMap()[r]
	return name, ok
}

// getCodeFromKeyName returns the Unicode code point for a key name
func getCodeFromKeyName(name string) (rune, bool) {
	for code, n := range xcodeKeyCodeMap() {
		if n == name {
			return code, true
		}
	}
	return 0, false
}
