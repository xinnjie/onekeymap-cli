package metrics_test

import (
	"testing"

	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestUnknownActionReporter_ReportUnknownCommand(t *testing.T) {
	ctx := t.Context()

	reporter := metrics.NewUnknownActionReporter(metrics.NewNoop())

	reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeVSCode, "workbench.action.unknown")
	reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeIntelliJ, "$Cut")
	reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeZed, "editor::Cut")
}
