package xcode

import (
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
			if xc.TextAction.TextAction == textAction && !xc.DisableImport {
				return &mapping
			}
		}
	}
	return nil
}
