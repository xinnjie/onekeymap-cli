package xcode

import (
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

// FindByXcodeAction searches for a mapping by Xcode action.
func (i *xcodeImporter) FindByXcodeAction(action string) *mappings.ActionMappingConfig {
	for _, mapping := range i.mappingConfig.Mappings {
		for _, xc := range mapping.Xcode {
			if xc.Action == action && !xc.DisableImport {
				return &mapping
			}
		}
	}
	return nil
}
