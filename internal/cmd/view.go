package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/views"
)

var (
	viewFile *string
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current OneKeymapSetting in a read-only TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := *viewFile
		if path == "" {
			path = viper.GetString("onekeymap")
		}
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open onekeymap config: %w", err)
		}
		defer func() { _ = f.Close() }()

		setting, err := keymap.Load(f)
		if err != nil {
			return fmt.Errorf("failed to parse onekeymap config: %w", err)
		}

		m := views.NewKeymapViewModel(setting, mappingConfig)
		_, err = tea.NewProgram(m).Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewFile = viewCmd.Flags().String("file", "", "Path to onekeymap.json (defaults to config value)")
}
