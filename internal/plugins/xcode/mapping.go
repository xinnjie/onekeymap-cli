package xcode

import (
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

// FindByXcodeAction searches for a mapping by Xcode action.
func (i *xcodeImporter) FindByXcodeAction(action string) *mappings.ActionMappingConfig {
	for _, mapping := range i.mappingConfig.Mappings {
		for _, xc := range mapping.Xcode {
			if xc.MenuAction.Action == action && !xc.DisableImport {
				return &mapping
			}
		}
	}
	return nil
}

// FindByXcodeTextAction searches for a mapping by Xcode text action.
func (i *xcodeImporter) FindByXcodeTextAction(textAction string) *mappings.ActionMappingConfig {
	for _, mapping := range i.mappingConfig.Mappings {
		for _, xc := range mapping.Xcode {
			if xc.DisableImport {
				continue
			}
			items := xc.TextAction.TextAction.Items
			if len(items) == 0 {
				continue
			}
			// Skip multi-action mappings during import
			if len(items) > 1 {
				continue
			}
			for _, a := range items {
				if a == "" {
					continue
				}
				if !strings.HasSuffix(a, ":") {
					a += ":"
				}
				if a == textAction {
					return &mapping
				}
			}
		}
	}
	return nil
}
