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
	importMetric := Metric{
		Name:        "import_total",
		Unit:        "1",
		Description: "Total number of import operations called",
	}

	exportMetric := Metric{
		Name:        "export_total",
		Unit:        "1",
		Description: "Total number of export operations called",
	}

	return &ServiceReporter{
		importCounter: recorder.Counter(importMetric),
		exportCounter: recorder.Counter(exportMetric),
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
