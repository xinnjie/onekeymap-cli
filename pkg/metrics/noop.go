package metrics

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// noopRecorder implements the Recorder interface but does nothing.
type noopRecorder struct {
	meter otelmetric.Meter
}

// Meter returns a no-op meter.
func (r *noopRecorder) Meter() otelmetric.Meter {
	return r.meter
}

// Provider returns a nil provider for noop recorder.
// Note: This might need to return a real but empty provider if callers rely on it being non-nil.
func (r *noopRecorder) Provider() *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider()
}

// Shutdown is a no-op.
func (r *noopRecorder) Shutdown(_ context.Context) error { return nil }

// NewNoop returns a no-op Recorder.
func NewNoop() Recorder {
	return &noopRecorder{
		meter: noop.NewMeterProvider().Meter("noop"),
	}
}
