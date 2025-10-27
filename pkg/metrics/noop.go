package metrics

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// noopRecorder implements the Recorder interface but does nothing.
type noopRecorder struct{}

// Histogram returns a no-op histogram.
func (r *noopRecorder) Histogram(_ Metric) otelmetric.Int64Histogram { //nolint:ireturn
	return noop.Int64Histogram{}
}

// Counter returns a no-op counter.
func (r *noopRecorder) Counter(_ Metric) otelmetric.Int64Counter { //nolint:ireturn
	return noop.Int64Counter{}
}

// Shutdown is a no-op.
func (r *noopRecorder) Shutdown(_ context.Context) error { return nil }

// NewNoop returns a no-op Recorder.
func NewNoop() Recorder {
	return &noopRecorder{}
}
