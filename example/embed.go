package example

import (
	"embed"
)

// ConfigFS contains the embedded example config template
//
//go:embed config.yaml
var ConfigFS embed.FS
