package metrics

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
)

// ServiceReporter handles metrics reporting for import and export service operations.
type ServiceReporter struct {
	importCounter otelmetric.Int64Counter
	exportCounter otelmetric.Int64Counter
}

// NewServiceReporter creates a new metrics reporter for service operations.
func NewServiceReporter(recorder Recorder) *ServiceReporter {
	meter := recorder.Meter()

	importCounter, _ := meter.Int64Counter(
		"onekeymap.import_total",
		otelmetric.WithDescription("Total number of import operations called"),
		otelmetric.WithUnit("1"),
	)

	exportCounter, _ := meter.Int64Counter(
		"onekeymap.export_total",
		otelmetric.WithDescription("Total number of export operations called"),
		otelmetric.WithUnit("1"),
	)

	return &ServiceReporter{
		importCounter: importCounter,
		exportCounter: exportCounter,
	}
}

// ReportImportCall reports an import operation.
func (r *ServiceReporter) ReportImportCall(ctx context.Context) {
	if r.importCounter != nil {
		r.importCounter.Add(ctx, 1)
	}
}

// ReportExportCall reports an export operation.
func (r *ServiceReporter) ReportExportCall(ctx context.Context) {
	if r.exportCounter != nil {
		r.exportCounter.Add(ctx, 1)
	}
}
