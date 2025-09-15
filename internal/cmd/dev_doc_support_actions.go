package cmd

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
)

// docSupportActionsCmd represents the docSupportActions command.
var docSupportActionsCmd = &cobra.Command{
	Use:   "docSupportActions",
	Short: "Generate markdown table showing action support across editors",
	Long: `Reads all action mappings and generates a markdown table showing which editors
support each action. The table includes columns for VSCode, Zed, IntelliJ, and Helix.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load all mappings
		mappingConfig, err := mappings.NewMappingConfig()
		if err != nil {
			log.Fatalf("Error loading mapping config: %v", err)
		}

		// Collect all action IDs and sort them
		actionIDs := make([]string, 0, len(mappingConfig.Mappings))
		for id := range mappingConfig.Mappings {
			actionIDs = append(actionIDs, id)
		}
		sort.Strings(actionIDs)

		// Generate markdown table
		fmt.Println("# Action Support Matrix")
		fmt.Println()
		fmt.Println("| Action | VSCode | Zed | IntelliJ | Helix | Description | Action ID |")
		fmt.Println("|--------|--------|-----|----------|-------|-------------|-----------|")

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

			fmt.Printf("| %s | %s | %s | %s | %s | %s | %s |\n",
				mapping.Name,
				formatSupport(vscodeSupport, vscodeReason),
				formatSupport(zedSupport, zedReason),
				formatSupport(intellijSupport, intellijReason),
				formatSupport(helixSupport, helixReason),
				description,
				id,
			)
		}
	},
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

func init() {
	devCmd.AddCommand(docSupportActionsCmd)
}
