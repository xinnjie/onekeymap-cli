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
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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
			Action         string
			VSCode         string
			Zed            string
			IntelliJ       string
			Xcode          string
			Helix          string
			Description    string
			ActionID       string
			FeaturedReason string
		}

		type categorySection struct {
			Category     string
			Rows         []supportRow
			FeaturedRows []supportRow
		}

		// Group actions by category, separating common and featured
		categoryMap := make(map[string][]supportRow)
		featuredCategoryMap := make(map[string][]supportRow)
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

			// Format featured reason for markdown (escape pipes and newlines)
			featuredReason := strings.ReplaceAll(mapping.FeaturedReason, "|", "\\|")
			featuredReason = strings.ReplaceAll(featuredReason, "\n", " ")
			if featuredReason == "" {
				featuredReason = "-"
			}

			row := supportRow{
				Action:         mapping.Name,
				VSCode:         formatSupport(vscodeSupport, vscodeReason),
				Zed:            formatSupport(zedSupport, zedReason),
				IntelliJ:       formatSupport(intellijSupport, intellijReason),
				Xcode:          formatSupport(xcodeSupport, xcodeReason),
				Helix:          formatSupport(helixSupport, helixReason),
				Description:    description,
				ActionID:       id,
				FeaturedReason: featuredReason,
			}

			// Separate common and featured actions
			if mapping.Featured {
				featuredCategoryMap[category] = append(featuredCategoryMap[category], row)
			} else {
				categoryMap[category] = append(categoryMap[category], row)
			}
		}

		// Sort categories and rows within each category
		// Collect all categories (from both common and featured)
		allCategories := make(map[string]bool)
		for category := range categoryMap {
			allCategories[category] = true
		}
		for category := range featuredCategoryMap {
			allCategories[category] = true
		}

		categories := make([]string, 0, len(allCategories))
		for category := range allCategories {
			categories = append(categories, category)
		}
		sort.Strings(categories)

		// Build category sections
		sections := make([]categorySection, 0, len(categories))
		for _, category := range categories {
			rows := categoryMap[category]
			featuredRows := featuredCategoryMap[category]

			// Sort rows by ActionID within each category
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].ActionID < rows[j].ActionID
			})
			sort.Slice(featuredRows, func(i, j int) bool {
				return featuredRows[i].ActionID < featuredRows[j].ActionID
			})

			sections = append(sections, categorySection{
				Category:     category,
				Rows:         rows,
				FeaturedRows: featuredRows,
			})
		}

		const supportMatrixTmpl = `# Action Support Matrix
{{- range . }}

## {{ .Category }}
{{- if .Rows }}

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
{{- range .Rows }}
| {{ .Action }} | {{ .VSCode }} | {{ .Zed }} | {{ .IntelliJ }} | {{ .Xcode }} | {{ .Helix }} | {{ .Description }} | {{ .ActionID }} |
{{- end }}
{{- end }}
{{- if .FeaturedRows }}

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
{{- range .FeaturedRows }}
| {{ .Action }} | {{ .VSCode }} | {{ .Zed }} | {{ .IntelliJ }} | {{ .Xcode }} | {{ .Helix }} | {{ .Description }} | {{ .ActionID }} | {{ .FeaturedReason }} |
{{- end }}
</details>

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
		if reason != "" {
			return fmt.Sprintf("✅ (%s)", reason)
		}
		return "✅"
	}
	if reason == "__explicitly_not_supported__" {
		return "❌"
	}
	if reason != "" {
		return fmt.Sprintf("❌ (%s)", reason)
	}
	return "N/A"
}
