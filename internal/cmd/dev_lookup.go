package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/keybindinglookup"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

type devLookupFlags struct {
	editor     string
	keybind    string
	configPath string
}

func NewCmdDevLookup() *cobra.Command {
	f := devLookupFlags{}

	// Create and configure the lookup factory
	factory := keybindinglookup.NewLookupFactory()
	factory.Register("vscode", vscode.NewVSCodeKeybindingLookup)
	factory.Register("xcode", xcode.NewXcodeKeybindingLookup)

	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Lookup keybindings in editor-specific configuration",
		Long: `Lookup keybindings in editor-specific configuration files.
Supports ` + factory.GetSupportedEditors() + ` editors.

Examples:
  # Lookup cmd+k in VSCode keybindings
  onekeymap-cli dev lookup --editor vscode --keybind "cmd+k" --config path/to/keybindings.json

  # Lookup @k in Xcode keybindings
  onekeymap-cli dev lookup --editor xcode --keybind "cmd+k" --config path/to/keybindings.idekeybindings`,
		Run:  devLookupRun(&f, factory),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.editor, "editor", "", "Editor type ("+factory.GetSupportedEditors()+")")
	cmd.Flags().StringVar(&f.keybind, "keybind", "", "Key binding to lookup (e.g., cmd+k, ctrl+shift+p)")
	cmd.Flags().StringVar(&f.configPath, "config", "", "Path to editor-specific configuration file")

	if err := cmd.MarkFlagRequired("editor"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("keybind"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}

	return cmd
}

func devLookupRun(f *devLookupFlags, factory *keybindinglookup.LookupFactory) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		// Parse the keybinding
		keybinding, err := keybinding.NewKeybinding(f.keybind, keybinding.ParseOption{Separator: "+"})
		if err != nil {
			cmd.PrintErrf("Error parsing keybinding '%s': %v\n", f.keybind, err)
			os.Exit(1)
		}

		// Open the configuration file
		configFile, err := os.Open(f.configPath)
		if err != nil {
			cmd.PrintErrf("Error opening config file '%s': %v\n", f.configPath, err)
			os.Exit(1)
		}
		defer configFile.Close()

		// Create lookup implementation using factory
		lookup, err := factory.CreateLookup(f.editor)
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		// Perform the lookup
		results, err := lookup.Lookup(configFile, keybinding)
		if err != nil {
			cmd.PrintErrf("Error looking up keybinding in %s config: %v\n", f.editor, err)
			os.Exit(1)
		}

		// Display results
		if len(results) == 0 {
			cmd.Printf("No matching keybindings found for '%s' in %s configuration.\n", f.keybind, f.editor)
			return
		}

		cmd.Printf(
			"Found %d matching keybinding(s) for '%s' in %s configuration:\n\n",
			len(results),
			f.keybind,
			f.editor,
		)

		for i, result := range results {
			cmd.Printf("Match %d:\n", i+1)

			// Pretty print JSON
			var prettyJSON map[string]interface{}
			if err := json.Unmarshal([]byte(result), &prettyJSON); err != nil {
				cmd.Printf("  Raw JSON: %s\n", result)
			} else {
				prettyBytes, err := json.MarshalIndent(prettyJSON, "  ", "  ")
				if err != nil {
					cmd.Printf("  Raw JSON: %s\n", result)
				} else {
					cmd.Printf("  %s\n", string(prettyBytes))
				}
			}
			cmd.Println()
		}
	}
}
