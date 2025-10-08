package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

type devDocSupportActionsFlags struct {
	// No flags for this command currently
}

func NewCmdDevDocSupportActions() *cobra.Command {
	f := devDocSupportActionsFlags{}
	cmd := &cobra.Command{
		Use:   "docSupportActions",
		Short: "Generate markdown table showing action support across editors",
		Long: `Reads all action mappings and generates a markdown table showing which editors
support each action. The table includes columns for VSCode, Zed, IntelliJ, and Helix.`,
		Run: devDocSupportActionsRun(&f, func() *slog.Logger {
			return cmdLogger
		}),
		Args: cobra.ExactArgs(0),
	}

	return cmd
}

func devDocSupportActionsRun(
	_ *devDocSupportActionsFlags,
	dependencies func() *slog.Logger,
) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger := dependencies()
		ctx := cmd.Context()
		// Load all mappings
		mappingConfig, err := mappings.NewMappingConfig()
		if err != nil {
			logger.ErrorContext(ctx, "Error loading mapping config", "error", err)
			os.Exit(1)
		}

		// Collect all action IDs and sort them
		actionIDs := make([]string, 0, len(mappingConfig.Mappings))
		for id := range mappingConfig.Mappings {
			actionIDs = append(actionIDs, id)
		}
		sort.Strings(actionIDs)

		// Prepare data rows for template rendering
		type supportRow struct {
			Action      string
			VSCode      string
			Zed         string
			IntelliJ    string
			Helix       string
			Description string
			ActionID    string
		}

		rows := make([]supportRow, 0, len(actionIDs))
		for _, id := range actionIDs {
			mapping := mappingConfig.Mappings[id]

			// Check support for each editor
			vscodeSupport, vscodeReason := checkVSCodeSupport(mapping)
			zedSupport, zedReason := checkZedSupport(mapping)
			intellijSupport, intellijReason := checkIntelliJSupport(mapping)
			helixSupport, helixReason := checkHelixSupport(mapping)

			// Format description for markdown (escape pipes and newlines)
			description := strings.ReplaceAll(mapping.Description, "|", "\\|")
			description = strings.ReplaceAll(description, "\n", " ")
			if description == "" {
				description = "-"
			}

			rows = append(rows, supportRow{
				Action:      mapping.Name,
				VSCode:      formatSupport(vscodeSupport, vscodeReason),
				Zed:         formatSupport(zedSupport, zedReason),
				IntelliJ:    formatSupport(intellijSupport, intellijReason),
				Helix:       formatSupport(helixSupport, helixReason),
				Description: description,
				ActionID:    id,
			})
		}

		const supportMatrixTmpl = `# Action Support Matrix

| Action | VSCode | Zed | IntelliJ | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------------|-----------|
{{- range . }}
| {{ .Action }} | {{ .VSCode }} | {{ .Zed }} | {{ .IntelliJ }} | {{ .Helix }} | {{ .Description }} | {{ .ActionID }} |
{{- end }}
`

		t := template.Must(template.New("support-matrix").Parse(supportMatrixTmpl))
		if err := t.Execute(cmd.OutOrStdout(), rows); err != nil {
			logger.ErrorContext(ctx, "Error executing template", "error", err)
			os.Exit(1)
		}
	}
}

func checkVSCodeSupport(mapping mappings.ActionMappingConfig) (bool, string) {
	// Check if explicitly marked as not supported
	for _, vc := range mapping.VSCode {
		if vc.NotSupported {
			return false, vc.NotSupportedReason
		}
		if vc.Command != "" {
			return true, ""
		}
	}
	return false, ""
}

func checkZedSupport(mapping mappings.ActionMappingConfig) (bool, string) {
	// Check if explicitly marked as not supported
	for _, zc := range mapping.Zed {
		if zc.NotSupported {
			return false, zc.NotSupportedReason
		}
		if zc.Action != "" {
			return true, ""
		}
	}
	return false, ""
}

func checkIntelliJSupport(mapping mappings.ActionMappingConfig) (bool, string) {
	if mapping.IntelliJ.NotSupported {
		return false, mapping.IntelliJ.NotSupportedReason
	}
	return mapping.IntelliJ.Action != "", ""
}

func checkHelixSupport(mapping mappings.ActionMappingConfig) (bool, string) {
	// Check if explicitly marked as not supported
	for _, hc := range mapping.Helix {
		if hc.NotSupported {
			return false, hc.NotSupportedReason
		}
		if hc.Command != "" {
			return true, ""
		}
	}
	return false, ""
}

func formatSupport(supported bool, reason string) string {
	if supported {
		return "✅"
	}
	if reason != "" {
		return fmt.Sprintf("❌ (%s)", reason)
	}
	return "N/A"
}
