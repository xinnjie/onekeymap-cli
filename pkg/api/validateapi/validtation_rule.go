package validateapi

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type ValidationRule interface {
	Validate(ctx context.Context, validationContext *ValidationContext) error
}

// ValidationContext holds all the necessary data for a validation rule to execute.
type ValidationContext struct {
	Setting keymap.Keymap
	Report  *keymapv1.ValidationReport
	Options importerapi.ImportOptions
}
