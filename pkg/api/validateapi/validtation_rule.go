package validateapi

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type ValidationRule interface {
	Validate(ctx context.Context, validationContext *ValidationContext) error
}

// ValidationContext holds all the necessary data for a validation rule to execute.
type ValidationContext struct {
	Setting    keymap.Keymap
	Report     *ValidationReport
	EditorType pluginapi.EditorType
}
