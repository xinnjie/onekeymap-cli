package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

type ActionDetailsViewModel struct {
	actionID      string
	description   string
	category      string
	editorSupport map[pluginapi.EditorType]editorSupportInfo
	mappingConfig *mappings.ActionMappingConfig
}

type editorSupportInfo struct {
	supported          bool
	notSupportedReason string
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
	// Define all main editor types to check
	editorTypes := []pluginapi.EditorType{
		pluginapi.EditorTypeVSCode,
		pluginapi.EditorTypeIntelliJ,
		pluginapi.EditorTypeZed,
		pluginapi.EditorTypeVim,
		pluginapi.EditorTypeHelix,
	}

	for _, editorType := range editorTypes {
		supported, notSupportedReason := mapping.IsSupported(editorType)
		info := editorSupportInfo{
			supported:          supported,
			notSupportedReason: notSupportedReason,
		}

		d.editorSupport[editorType] = info
	}
}

func (d *ActionDetailsViewModel) View() string {
	var b strings.Builder

	labelStyle := lipgloss.NewStyle().Bold(true)

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
		b.WriteString(d.renderEditorSupport(labelStyle))
	}

	return b.String()
}

func (d *ActionDetailsViewModel) renderEditorSupport(labelStyle lipgloss.Style) string {
	var b strings.Builder

	supportedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))   // Green
	notSupportedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red

	b.WriteString(labelStyle.Render("Editor Support:") + "\n")

	// Define display order
	editorOrder := []pluginapi.EditorType{
		pluginapi.EditorTypeVSCode,
		pluginapi.EditorTypeIntelliJ,
		pluginapi.EditorTypeZed,
		pluginapi.EditorTypeVim,
		pluginapi.EditorTypeHelix,
	}

	for _, editorType := range editorOrder {
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
	if info.notSupportedReason != "" {
		status += fmt.Sprintf(" (%s)", info.notSupportedReason)
	}
	return status
}
