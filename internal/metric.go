package internal

import "github.com/xinnjie/onekeymap-cli/pkg/metrics"

var (
	// nolint: gochecknoglobals // metricImportCalls measures the number of import operations called.
	metricImportCalls = metrics.Metric{
		Name:        "import_calls",
		Unit:        "{count}",
		Description: "Measures the number of import operations called.",
	}

	// nolint: gochecknoglobals // metricExportCalls measures the number of export operations called.
	metricExportCalls = metrics.Metric{
		Name:        "export_calls",
		Unit:        "{count}",
		Description: "Measures the number of export operations called.",
	}
)
