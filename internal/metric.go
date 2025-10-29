package internal

import "github.com/xinnjie/onekeymap-cli/internal/metrics"

var (
	// nolint: gochecknoglobals // metricImportCalls measures the number of import operations called.
	metricImportCalls = metrics.Metric{
		Name:        "import",
		Unit:        "{count}",
		Description: "Measures the number of import operations called.",
	}

	// nolint: gochecknoglobals // metricExportCalls measures the number of export operations called.
	metricExportCalls = metrics.Metric{
		Name:        "export",
		Unit:        "{count}",
		Description: "Measures the number of export operations called.",
	}
)
