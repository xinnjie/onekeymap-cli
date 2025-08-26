package intellij

import "encoding/xml"

// XML models for IntelliJ keymap parsing/exporting.

type KeymapXML struct {
	XMLName          xml.Name    `xml:"keymap"`
	Name             string      `xml:"name,attr"`
	Version          string      `xml:"version,attr"`
	DisableMnemonics bool        `xml:"disable-mnemonics,attr"`
	Actions          []ActionXML `xml:"action"`
	Parent           string      `xml:"parent,attr"`
}

type ActionXML struct {
	ShortcutXML
	ID string `xml:"id,attr"`
}

type ShortcutXML struct {
	KeyboardShortcuts []KeyboardShortcutXML `xml:"keyboard-shortcut"`
	MouseShortcuts    []MouseShortcutXML    `xml:"mouse-shortcut"`
}

type KeyboardShortcutXML struct {
	First  string `xml:"first-keystroke,attr"`
	Second string `xml:"second-keystroke,attr,omitempty"`
}

type MouseShortcutXML struct {
	Keystroke string `xml:"keystroke,attr"`
}
