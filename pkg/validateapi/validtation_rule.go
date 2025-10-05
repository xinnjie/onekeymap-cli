package validateapi

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type ValidationRule interface {
	Validate(ctx context.Context, validationContext *ValidationContext) error
}

// ValidationContext holds all the necessary data for a validation rule to execute.
type ValidationContext struct {
	Setting *keymapv1.Keymap
	Report  *keymapv1.ValidationReport
	Options importapi.ImportOptions
}
