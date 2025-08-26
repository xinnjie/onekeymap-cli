/*
Copyright 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"slices"
	"sort"

	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
)

var editor string

// listUnmappedActionsCmd represents the generic unmapped actions command
var listUnmappedActionsCmd = &cobra.Command{
	Use:     "listUnmappedActions",
	Aliases: []string{"vscodeListUnmappedCommands"},
	Short:   "List action IDs without mappings for a given editor.",
	Long:    `Reads all action mappings and prints the list of action IDs that do not have a mapping for the specified editor (vscode|intellij|zed|vim|helix).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load all mappings
		mappingConfig, err := mappings.NewMappingConfig()
		if err != nil {
			log.Fatalf("Error loading mapping config: %v", err)
		}

		unmapped := make([]string, 0)
		for id, m := range mappingConfig.Mappings {
			mapped := false
			// Skip entries explicitly marked as notSupported for the selected editor
			switch editor {
			case "vscode":
				skip := false
				for _, vc := range m.VSCode {
					if vc.NotSupported {
						skip = true
						break
					}
				}
				if skip {
					continue
				}
			case "intellij":
				if m.IntelliJ.NotSupported {
					continue
				}
			case "zed":
				skip := slices.ContainsFunc(m.Zed, func(zc mappings.ZedMappingConfig) bool {
					return zc.NotSupported
				})
				if skip {
					continue
				}
			case "vim":
				if m.Vim.NotSupported {
					continue
				}
			case "helix":
				skip := slices.ContainsFunc(m.Helix, func(hc mappings.HelixMappingConfig) bool {
					return hc.NotSupported
				})
				if skip {
					continue
				}
			default:
				log.Fatalf("Unknown editor: %s (valid: vscode|intellij|zed|vim|helix)", editor)
			}

			switch editor {
			case "vscode":
				for _, vc := range m.VSCode {
					if vc.Command != "" {
						mapped = true
						break
					}
				}
			case "intellij":
				if m.IntelliJ.Action != "" {
					mapped = true
				}
			case "zed":
				for _, zc := range m.Zed {
					if zc.Action != "" {
						mapped = true
						break
					}
				}
			case "vim":
				if m.Vim.Command != "" {
					mapped = true
				}
			case "helix":
				for _, hc := range m.Helix {
					if hc.Command != "" {
						mapped = true
						break
					}
				}
			}
			if !mapped {
				unmapped = append(unmapped, id)
			}
		}
		sort.Strings(unmapped)
		fmt.Printf("Unmapped action IDs for editor %q:\n", editor)
		for _, id := range unmapped {
			fmt.Println(id)
		}
	},
}

func init() {
	devCmd.AddCommand(listUnmappedActionsCmd)
	listUnmappedActionsCmd.Flags().StringVar(&editor, "editor", "vscode", "Editor to check: vscode|intellij|zed|vim|helix")
}
