package views

import (
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

type ActionDetailsViewModel struct {
	actionID    string
	description string
	category    string
}

func newActionDetailsViewModel(actionID string, mc *mappings.MappingConfig) ActionDetailsViewModel {
	d := ActionDetailsViewModel{actionID: actionID}
	if mc == nil {
		return d
	}
	if mapping := mc.FindByUniversalAction(actionID); mapping != nil {
		if mapping.Name != "" {
			d.description = mapping.Name
		} else if mapping.Description != "" {
			d.description = mapping.Description
		}
		if mapping.Category != "" {
			d.category = mapping.Category
		}
	}
	return d
}

func (d ActionDetailsViewModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	fmt.Fprintf(&b, "Action: %s\n", d.actionID)
	if d.description != "" {
		fmt.Fprintf(&b, "Description: %s\n", d.description)
	}
	if d.category != "" {
		fmt.Fprintf(&b, "Category: %s\n", d.category)
	}
	return b.String()
}
