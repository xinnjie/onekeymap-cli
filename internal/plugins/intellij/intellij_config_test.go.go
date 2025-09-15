package intellij

import (
	"encoding/xml"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalKeymapXML(t *testing.T) {
	xmlData := `
<keymap name="$default" version="1" disable-mnemonics="false">
  <action id="ShowNavBar">
    <keyboard-shortcut first-keystroke="alt HOME"/>
  </action>
  <action id="OpenInRightSplit">
    <keyboard-shortcut first-keystroke="shift ENTER"/>
    <mouse-shortcut keystroke="alt button1 doubleClick"/>
  </action>
  <action id="GotoChangedFile"/>
</keymap>`

	var km KeymapXML
	err := xml.Unmarshal([]byte(xmlData), &km)
	require.NoError(t, err)

	assert.Equal(t, "$default", km.Name)
	assert.Equal(t, "1", km.Version)
	assert.False(t, km.DisableMnemonics)
	assert.Len(t, km.Actions, 3)

	showNavBarIndex := slices.IndexFunc(km.Actions, func(a ActionXML) bool {
		return a.ID == "ShowNavBar"
	})
	require.GreaterOrEqual(t, showNavBarIndex, 0)

	showNavBar := km.Actions[showNavBarIndex]
	require.NotNil(t, showNavBar)
	assert.Len(t, showNavBar.KeyboardShortcuts, 1)
	assert.Equal(t, "alt HOME", showNavBar.KeyboardShortcuts[0].First)
	assert.Empty(t, showNavBar.KeyboardShortcuts[0].Second)

	openRightIndex := slices.IndexFunc(km.Actions, func(a ActionXML) bool {
		return a.ID == "OpenInRightSplit"
	})
	require.GreaterOrEqual(t, openRightIndex, 0)

	openRight := km.Actions[openRightIndex]
	require.NotNil(t, openRight)
	assert.Len(t, openRight.KeyboardShortcuts, 1)
	assert.Equal(t, "shift ENTER", openRight.KeyboardShortcuts[0].First)
	assert.Len(t, openRight.MouseShortcuts, 1)
	assert.Equal(t, "alt button1 doubleClick", openRight.MouseShortcuts[0].Keystroke)
}
