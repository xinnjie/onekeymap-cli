package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync keymaps directly between two editors",
	RunE: func(cmd *cobra.Command, args []string) error {

		interactive, _ := cmd.Flags().GetBool("interactive")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		inputFile, _ := cmd.Flags().GetString("input")
		outputFile, _ := cmd.Flags().GetString("output")

		if interactive || from == "" || to == "" {
			logger.Info("Interactive mode for sync is not yet implemented.")
			return nil
		}

		logger.Info("Syncing keymaps", "from", from, "to", to)

		inputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(from))
		if !ok {
			logger.Error("failed to get input plugin", "from", from)
			return fmt.Errorf("failed to get input plugin")
		}
		outputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(to))
		if !ok {
			logger.Error("failed to get output plugin", "to", to)
			return fmt.Errorf("failed to get output plugin")
		}

		if inputFile == "" {
			v, err := inputPlugin.DefaultConfigPath()
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			inputFile = v[0]
		}

		if outputFile == "" {
			v, err := outputPlugin.DefaultConfigPath()
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			outputFile = v[0]
		}

		inputStream, err := os.Open(inputFile)
		if err != nil {
			logger.Error("failed to open input file", "error", err)
			return err
		}
		defer func() { _ = inputStream.Close() }()

		importOpts := importapi.ImportOptions{
			EditorType:  pluginapi.EditorType(from),
			InputStream: inputStream,
		}
		importResult, err := importService.Import(cmd.Context(), importOpts)
		if err != nil {
			logger.Error("sync failed during import step", "error", err)
			return err
		}

		logger.Info("Import complete.")
		if importResult != nil {
			logger.Debug("Import Report", "report", importResult.Report)
		}

		if importResult == nil || importResult.Setting == nil {
			logger.Warn("No imported keymaps to export; aborting sync")
			return nil
		}

		outputStream, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			logger.Error("failed to create output file", "error", err)
			return err
		}
		defer func() { _ = outputStream.Close() }()

		exportOpts := exportapi.ExportOptions{EditorType: pluginapi.EditorType(to)}
		exportReport, err := exportService.Export(outputStream, importResult.Setting, exportOpts)
		if err != nil {
			logger.Error("sync failed during export step", "error", err)
			return err
		}

		logger.Info("Export complete.")
		logger.Debug("Export Report", "report", exportReport)
		logger.Info("Sync complete!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().String("from", "", "Source editor, valid values: vscode, zed")
	syncCmd.Flags().String("to", "", "Target editor, valid values: vscode, zed")
	syncCmd.Flags().String("input", "", "Path to source editor config")
	syncCmd.Flags().String("output", "", "Path to target editor config")

	syncCmd.Flags().Bool("backup", false, "Create a backup of the target editor's keymap")
	_ = syncCmd.MarkFlagRequired("from")
	_ = syncCmd.MarkFlagRequired("to")
}
