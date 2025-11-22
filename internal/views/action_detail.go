package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

// nolint:gochecknoglobals // editorTypesDisplayOrder defines the order in which editor types are displayed
var editorTypesDisplayOrder = [...]pluginapi.EditorType{
	pluginapi.EditorTypeVSCode,
	pluginapi.EditorTypeIntelliJ,
	pluginapi.EditorTypeZed,
	pluginapi.EditorTypeHelix,
}

type ActionDetailsViewModel struct {
	actionID      string
	description   string
	category      string
	editorSupport map[pluginapi.EditorType]editorSupportInfo
	mappingConfig *mappings.ActionMappingConfig
}

type editorSupportInfo struct {
	supported bool
	note      string
}

func newActionDetailsViewModel(actionID string, mc *mappings.MappingConfig) *ActionDetailsViewModel {
	d := &ActionDetailsViewModel{
		actionID:      actionID,
		editorSupport: make(map[pluginapi.EditorType]editorSupportInfo),
	}
	if mc == nil {
		return d
	}
	mapping := mc.Get(actionID)
	if mapping == nil {
		return d
	}

	d.mappingConfig = mapping

	if mapping.Name != "" {
		d.description = mapping.Name
	} else if mapping.Description != "" {
		d.description = mapping.Description
	}
	if mapping.Category != "" {
		d.category = mapping.Category
	}

	// Collect editor support information
	d.collectEditorSupport(mapping)

	return d
}

func (d *ActionDetailsViewModel) collectEditorSupport(mapping *mappings.ActionMappingConfig) {
	for _, editorType := range editorTypesDisplayOrder {
		supported, note := mapping.IsSupported(editorType)
		info := editorSupportInfo{
			supported: supported,
			note:      note,
		}

		d.editorSupport[editorType] = info
	}
}

func (d *ActionDetailsViewModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Action:"), d.actionID)
	if d.description != "" {
		fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Description:"), d.description)
	}
	if d.category != "" {
		fmt.Fprintf(&b, "%s %s\n", labelStyle.Render("Category:"), d.category)
	}

	// Display editor support
	if len(d.editorSupport) > 0 {
		b.WriteString("\n")
		b.WriteString(d.renderEditorSupport())
	}

	return b.String()
}

func (d *ActionDetailsViewModel) renderEditorSupport() string {
	var b strings.Builder

	b.WriteString(labelStyle.Render("Editor Support:") + "\n")

	for _, editorType := range editorTypesDisplayOrder {
		info, exists := d.editorSupport[editorType]
		if !exists {
			continue
		}

		editorName := editorType.AppName()
		status := d.formatEditorStatus(info, supportedStyle, notSupportedStyle)
		fmt.Fprintf(&b, "  %-20s %s\n", editorName+":", status)
	}

	return b.String()
}

func (d *ActionDetailsViewModel) formatEditorStatus(
	info editorSupportInfo,
	supportedStyle, notSupportedStyle lipgloss.Style,
) string {
	if info.supported {
		return supportedStyle.Render("✓ Supported")
	}

	status := notSupportedStyle.Render("✗ Not Supported")
	if info.note != "" {
		status += fmt.Sprintf(" (%s)", info.note)
	}
	return status
}
