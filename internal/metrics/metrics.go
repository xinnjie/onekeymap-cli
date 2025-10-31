package metrics

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/microsoft/go-deviceid"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	metricExportIntervalSeconds = 1
)

// Metric represents a metric that can be collected.
type Metric struct {
	Name        string
	Unit        string
	Description string
}

// Recorder is the interface for recording metrics.
type Recorder interface {
	Histogram(metric Metric) otelmetric.Int64Histogram
	Counter(metric Metric) otelmetric.Int64Counter
	Shutdown(ctx context.Context) error
}

// recorder implements the Recorder interface and sends metrics to an OTLP endpoint.
type recorder struct {
	logger   *slog.Logger
	provider *sdkmetric.MeterProvider
	meter    otelmetric.Meter
	reader   sdkmetric.Reader
	exporter sdkmetric.Exporter
}

// RecorderOption holds configuration for creating a Recorder.
type RecorderOption struct {
	Endpoint string
	Headers  map[string]string
	UseDelta bool
}

// New creates a new Recorder and initializes the OpenTelemetry provider.
func New(
	ctx context.Context,
	version string,
	logger *slog.Logger,
	option RecorderOption,
) (Recorder, error) {
	opts := []otlpmetrichttp.Option{}

	if option.Endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(option.Endpoint))
	}

	if len(option.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(option.Headers))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	deviceID := getDeviceID(logger)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("onekeymap-cli"),
			semconv.ServiceVersion(version),
			semconv.OSName(runtime.GOOS),
			semconv.ServiceInstanceID(deviceID),
		),
	)
	if err != nil {
		return nil, err
	}

	var reader sdkmetric.Reader

	// Only manual reader support delta temporality
	if option.UseDelta {
		manualOpts := []sdkmetric.ManualReaderOption{}
		manualOpts = append(
			manualOpts,
			sdkmetric.WithTemporalitySelector(func(_ sdkmetric.InstrumentKind) metricdata.Temporality {
				return metricdata.DeltaTemporality
			}),
		)
		reader = sdkmetric.NewManualReader(manualOpts...)
		logger.DebugContext(ctx, "Using manual reader for metric provider because delta temporality is set.")
	} else {
		reader = sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportIntervalSeconds*time.Second))
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	meter := provider.Meter("onekeymap-cli")

	recorder := &recorder{
		logger:   logger,
		provider: provider,
		meter:    meter,
		reader:   reader,
		exporter: exporter,
	}

	return recorder, nil
}

// Histogram creates a new int64 histogram metric.
func (r *recorder) Histogram(metric Metric) otelmetric.Int64Histogram { //nolint:ireturn
	histogram, err := r.meter.Int64Histogram(
		"onekeymap."+metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)

	if err != nil {
		r.logger.Warn("failed to create histogram", "error", err)
		return noop.Int64Histogram{}
	}

	return histogram
}

// Counter creates a new int64 up down counter metric.
func (r *recorder) Counter(metric Metric) otelmetric.Int64Counter { //nolint:ireturn
	counter, err := r.meter.Int64Counter(
		"onekeymap."+metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)

	if err != nil {
		r.logger.Warn("failed to create counter", "error", err)
		return noop.Int64Counter{}
	}
	return counter
}

// Shutdown shuts down the OpenTelemetry provider.
func (r *recorder) Shutdown(ctx context.Context) error {
	if err := r.collect(ctx); err != nil {
		r.logger.DebugContext(ctx, "failed to collect metrics during shutdown", "error", err)
	}

	return r.provider.Shutdown(ctx)
}

// collect triggers immediate collection and export
func (r *recorder) collect(ctx context.Context) error {
	manualReader, ok := r.reader.(*sdkmetric.ManualReader)
	if !ok {
		return nil
	}
	var rm metricdata.ResourceMetrics
	if err := manualReader.Collect(ctx, &rm); err != nil {
		r.logger.ErrorContext(ctx, "Failed to collect from manual reader", "error", err)
		return err
	}

	if err := r.exporter.Export(ctx, &rm); err != nil {
		r.logger.DebugContext(ctx, "Failed to export collected metrics", "error", err)
		return err
	}

	return nil
}

// getDeviceID returns a stable device identifier for this machine.
// It uses the github.com/microsoft/go-deviceid library to generate a consistent ID.
func getDeviceID(logger *slog.Logger) string {
	deviceID, err := deviceid.Get()
	if err != nil {
		logger.Warn("Failed to get device ID, using fallback", "error", err)
		// Fallback to a generic identifier
		return "unknown-device"
	}
	return deviceID
}
