package metrics

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	mappings "github.com/xinnjie/onekeymap-cli/internal/mappings"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// Metric names.
const (
	commandProcessedName = "onekeymap.import.commands.processed"
)

const metricExportIntervalSeconds = 15

// Attribute keys.
const (
	attrKeyEditor  = "editor"
	attrKeyCommand = "command"
	attrKeyMapped  = "mapped"
)

// Recorder is the interface for recording metrics.
type Recorder interface {
	RecordCommandProcessed(ctx context.Context, editor string, setting *keymapv1.Keymap)
	Shutdown(ctx context.Context) error
}

// recorder implements the Recorder interface and sends metrics to an OTLP endpoint.
type recorder struct {
	logger         *slog.Logger
	commandCounter metric.Int64Counter
	provider       *sdkmetric.MeterProvider
	mappingConfig  *mappings.MappingConfig
}

// noopRecorder implements the Recorder interface but does nothing.
type noopRecorder struct{}

// RecordCommandProcessed is a no-op.
func (r *noopRecorder) RecordCommandProcessed(_ context.Context, _ string, _ *keymapv1.Keymap) {
}

// Shutdown is a no-op.
func (r *noopRecorder) Shutdown(_ context.Context) error { return nil }

// NewNoop returns a no-op Recorder.
func NewNoop() Recorder {
	return &noopRecorder{}
}

// New creates a new Recorder and initializes the OpenTelemetry provider.
func New(
	ctx context.Context,
	version string,
	logger *slog.Logger,
	mappingConfig *mappings.MappingConfig,
) (Recorder, error) {
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "onekeymap-cli"),
			attribute.String("service.version", version),
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
	otel.SetMeterProvider(provider)

	meter := provider.Meter("onekeymap.import.service")
	commandCounter, err := meter.Int64Counter(
		commandProcessedName,
		metric.WithDescription("The number of keybinding commands processed during import."),
		metric.WithUnit("{command}"),
	)
	if err != nil {
		return nil, err
	}

	recorder := &recorder{
		logger:         logger,
		commandCounter: commandCounter,
		provider:       provider,
		mappingConfig:  mappingConfig,
	}

	return recorder, nil
}

// RecordCommandProcessed records that a command has been processed.
func (r *recorder) RecordCommandProcessed(ctx context.Context, editor string, setting *keymapv1.Keymap) {
	if r.commandCounter == nil || setting == nil {
		return
	}
	for _, binding := range setting.GetKeybindings() {
		isMapped := r.mappingConfig.IsActionMapped(binding.GetName())
		r.commandCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String(attrKeyEditor, editor),
			attribute.String(attrKeyCommand, binding.GetName()),
			attribute.Bool(attrKeyMapped, isMapped),
		))
	}
}

// Shutdown shuts down the OpenTelemetry provider.
func (r *recorder) Shutdown(ctx context.Context) error {
	if r.provider == nil {
		return nil
	}
	return r.provider.Shutdown(ctx)
}
