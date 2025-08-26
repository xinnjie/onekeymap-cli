package pluginapi

import "fmt"

var (
	// Returned when a plugin does not support a requested operation, like import or export.
	ErrNotSupported = fmt.Errorf("not supported")
)
