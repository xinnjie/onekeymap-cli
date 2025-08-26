package zed

import (
	"fmt"
	"reflect"
)

// actionIDFromZedWithArgs converts a Zed action, context, and args to a universal action ID.
func (p *zedImporter) actionIDFromZedWithArgs(action, context string, args map[string]interface{}) (string, error) {
	for _, mapping := range p.mappingConfig.Mappings {
		for _, zconf := range mapping.Zed {
			if zconf.Action == action && zconf.Context == context {
				if zconf.Args == nil && args == nil {
					return mapping.ID, nil
				} else if zconf.Args != nil && args != nil && reflect.DeepEqual(zconf.Args, args) {
					return mapping.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no mapping found for zed action: %s", action)
}
