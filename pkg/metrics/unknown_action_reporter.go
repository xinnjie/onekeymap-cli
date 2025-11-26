package metrics

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// UnknownActionReporter handles metrics reporting for unknown actions across all plugins.
type UnknownActionReporter struct {
	unknownActionCounter otelmetric.Int64Counter
}

// NewUnknownActionReporter creates a new metrics reporter for unknown actions.
func NewUnknownActionReporter(recorder Recorder) *UnknownActionReporter {
	meter := recorder.Meter()
	unknownActionCounter, _ := meter.Int64Counter(
		"onekeymap.unknown_action_total",
		otelmetric.WithDescription("Total number of unknown actions encountered during import"),
		otelmetric.WithUnit("1"),
	)
	return &UnknownActionReporter{
		unknownActionCounter: unknownActionCounter,
	}
}

// ReportUnknownCommand reports an unknown action encountered during import.
func (r *UnknownActionReporter) ReportUnknownCommand(
	ctx context.Context,
	editorType pluginapi.EditorType,
	action string,
) {
	if r.unknownActionCounter == nil {
		return
	}

	r.unknownActionCounter.Add(ctx, 1, otelmetric.WithAttributes(
		attribute.String("editor_type", string(editorType)),
		attribute.String("action", action),
	))
}
