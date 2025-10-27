package metrics

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc/credentials"
)

const (
	metricExportIntervalSeconds = 15
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
}

// RecorderOption holds configuration for creating a Recorder.
type RecorderOption struct {
	Endpoint string
	Headers  map[string]string
	Insecure bool
}

// New creates a new Recorder and initializes the OpenTelemetry provider.
func New(
	ctx context.Context,
	version string,
	logger *slog.Logger,
	option RecorderOption,
) (Recorder, error) {
	opts := []otlpmetricgrpc.Option{}

	if option.Endpoint != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(option.Endpoint))
	}

	if len(option.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(option.Headers))
	}

	if option.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	} else {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("onekeymap-cli"),
			semconv.ServiceVersion(version),
			semconv.OSName(runtime.GOOS),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportIntervalSeconds*time.Second)),
		),
	)

	meter := provider.Meter("onekeymap-cli")

	recorder := &recorder{
		logger:   logger,
		provider: provider,
		meter:    meter,
	}

	return recorder, nil
}

// Histogram creates a new int64 histogram metric.
func (r *recorder) Histogram(metric Metric) otelmetric.Int64Histogram { //nolint:ireturn
	histogram, err := r.meter.Int64Histogram(
		metric.Name,
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
		metric.Name,
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
	if r.provider == nil {
		return nil
	}
	return r.provider.Shutdown(ctx)
}
