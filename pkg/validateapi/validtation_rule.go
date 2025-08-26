package validateapi

import (
	"context"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type ValidationRule interface {
	Validate(ctx context.Context, validationContext *ValidationContext) error
}

// ValidationContext holds all the necessary data for a validation rule to execute.
type ValidationContext struct {
	Setting *keymapv1.KeymapSetting
	Report  *keymapv1.ValidationReport
	Options importapi.ImportOptions
}
