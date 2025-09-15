package pluginapi

import "errors"

var (
	// Returned when a plugin does not support a requested operation, like import or export.
	ErrNotSupported = errors.New("not supported")
)
