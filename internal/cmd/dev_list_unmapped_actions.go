package cmd

import (
	"log/slog"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
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
		Run: devListUnmappedActionsRun(&f, func() *slog.Logger {
			return cmdLogger
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.editor, "editor", "vscode", "Editor to check: vscode|intellij|zed|vim|helix")

	return cmd
}

func devListUnmappedActionsRun(
	f *devListUnmappedActionsFlags,
	dependencies func() *slog.Logger,
) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger := dependencies()
		ctx := cmd.Context()
		mappingConfig, err := mappings.NewMappingConfig()
		if err != nil {
			logger.ErrorContext(ctx, "Error loading mapping config", "error", err)
			os.Exit(1)
		}

		editorType := pluginapi.EditorType(f.editor)
		unmapped := make([]string, 0)
		for id, m := range mappingConfig.Mappings {
			supported, _ := m.IsSupported(editorType)
			if !supported {
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
