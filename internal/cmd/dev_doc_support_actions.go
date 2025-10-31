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
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
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
support each action. The table includes columns for VSCode, Zed, IntelliJ, Helix, and Xcode.`,
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

		// Prepare data structures for template rendering
		type supportRow struct {
			Action      string
			VSCode      string
			Zed         string
			IntelliJ    string
			Xcode       string
			Helix       string
			Description string
			ActionID    string
		}

		type categorySection struct {
			Category string
			Rows     []supportRow
		}

		// Group actions by category
		categoryMap := make(map[string][]supportRow)
		for id, mapping := range mappingConfig.Mappings {
			category := mapping.Category
			if category == "" {
				category = "Uncategorized"
			}

			// Check support for each editor
			vscodeSupport, vscodeReason := mapping.IsSupported(pluginapi.EditorTypeVSCode)
			zedSupport, zedReason := mapping.IsSupported(pluginapi.EditorTypeZed)
			intellijSupport, intellijReason := mapping.IsSupported(pluginapi.EditorTypeIntelliJ)
			xcodeSupport, xcodeReason := mapping.IsSupported(pluginapi.EditorTypeXcode)
			helixSupport, helixReason := mapping.IsSupported(pluginapi.EditorTypeHelix)

			// Format description for markdown (escape pipes and newlines)
			description := strings.ReplaceAll(mapping.Description, "|", "\\|")
			description = strings.ReplaceAll(description, "\n", " ")
			if description == "" {
				description = "-"
			}

			row := supportRow{
				Action:      mapping.Name,
				VSCode:      formatSupport(vscodeSupport, vscodeReason),
				Zed:         formatSupport(zedSupport, zedReason),
				IntelliJ:    formatSupport(intellijSupport, intellijReason),
				Xcode:       formatSupport(xcodeSupport, xcodeReason),
				Helix:       formatSupport(helixSupport, helixReason),
				Description: description,
				ActionID:    id,
			}

			categoryMap[category] = append(categoryMap[category], row)
		}

		// Sort categories and rows within each category
		categories := make([]string, 0, len(categoryMap))
		for category := range categoryMap {
			categories = append(categories, category)
		}
		sort.Strings(categories)

		// Build category sections
		sections := make([]categorySection, 0, len(categories))
		for _, category := range categories {
			rows := categoryMap[category]
			// Sort rows by ActionID within each category
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].ActionID < rows[j].ActionID
			})
			sections = append(sections, categorySection{
				Category: category,
				Rows:     rows,
			})
		}

		const supportMatrixTmpl = `# Action Support Matrix
{{- range . }}

## {{ .Category }}

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
{{- range .Rows }}
| {{ .Action }} | {{ .VSCode }} | {{ .Zed }} | {{ .IntelliJ }} | {{ .Xcode }} | {{ .Helix }} | {{ .Description }} | {{ .ActionID }} |
{{- end }}
{{- end }}
`

		t := template.Must(template.New("support-matrix").Parse(supportMatrixTmpl))
		if err := t.Execute(cmd.OutOrStdout(), sections); err != nil {
			logger.ErrorContext(ctx, "Error executing template", "error", err)
			os.Exit(1)
		}
	}
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
