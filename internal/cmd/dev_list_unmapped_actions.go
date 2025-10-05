package cmd

import (
	"os"
	"slices"
	"sort"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

type devListUnmappedActionsFlags struct {
	editor string
}

func NewCmdDevListUnmappedActions() *cobra.Command {
	f := devListUnmappedActionsFlags{}
	cmd := &cobra.Command{
		Use:     "listUnmappedActions",
		Aliases: []string{"vscodeListUnmappedCommands"},
		Short:   "List action IDs without mappings for a given editor.",
		Long:    `Reads all action mappings and prints the list of action IDs that do not have a mapping for the specified editor (vscode|intellij|zed|vim|helix).`,
		Run:     devListUnmappedActionsRun(&f),
		Args:    cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.editor, "editor", "vscode", "Editor to check: vscode|intellij|zed|vim|helix")

	return cmd
}

func devListUnmappedActionsRun(f *devListUnmappedActionsFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		// Load all mappings
		mappingConfig, err := mappings.NewMappingConfig()
		if err != nil {
			logger.ErrorContext(ctx, "Error loading mapping config", "error", err)
			os.Exit(1)
		}

		unmapped := make([]string, 0)
		for id, m := range mappingConfig.Mappings {
			mapped := false
			// Skip entries explicitly marked as notSupported for the selected editor
			switch f.editor {
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
				logger.WarnContext(ctx, "Unknown editor", "editor", f.editor)
			}

			switch f.editor {
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
		cmd.Println("Unmapped action IDs for editor", f.editor)
		for _, id := range unmapped {
			cmd.Println(id)
		}
	}
}
