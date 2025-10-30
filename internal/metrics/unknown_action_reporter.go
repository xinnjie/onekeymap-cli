package metrics

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// UnknownActionReporter handles metrics reporting for unknown actions across all plugins.
type UnknownActionReporter struct {
	unknownActionCounter otelmetric.Int64Counter
}

// NewUnknownActionReporter creates a new metrics reporter for unknown actions.
func NewUnknownActionReporter(recorder Recorder) *UnknownActionReporter {
	unknownActionMetric := Metric{
		Name:        "unknown_action_total",
		Unit:        "1",
		Description: "Total number of unknown actions encountered during import",
	}
	return &UnknownActionReporter{
		unknownActionCounter: recorder.Counter(unknownActionMetric),
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
