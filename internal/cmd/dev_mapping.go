package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/keybindinglookup"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/xcode"
)

type devMappingFlags struct {
	vscodeConfig string
	xcodeConfig  string
	outputFormat string
}

type KeybindingMapping struct {
	Keybinding  string `json:"keybinding"`
	VSCodeCmd   string `json:"vscode_command"`
	VSCodeWhen  string `json:"vscode_when,omitempty"`
	XcodeAction string `json:"xcode_action"`
	XcodeTitle  string `json:"xcode_title,omitempty"`
}

func NewCmdDevMapping() *cobra.Command {
	f := devMappingFlags{}

	// Create and configure the lookup factory
	factory := keybindinglookup.NewLookupFactory()
	factory.Register("vscode", vscode.NewVSCodeKeybindingLookup)
	factory.Register("xcode", xcode.NewXcodeKeybindingLookup)

	cmd := &cobra.Command{
		Use:   "mapping",
		Short: "Find VSCode command to Xcode action mappings by comparing keybindings",
		Long: `Compare VSCode keybindings.json and Xcode .idekeybindings files to find
command/action mappings based on shared keybindings.

This helps identify which VSCode commands correspond to which Xcode actions
by analyzing configurations where the same keybinding performs similar functions.

Examples:
  # Compare configurations and output as table
  onekeymap-cli dev mapping --vscode path/to/keybindings.json --xcode path/to/keybindings.idekeybindings

  # Output as JSON for further processing
  onekeymap-cli dev mapping --vscode path/to/keybindings.json --xcode path/to/keybindings.idekeybindings --format json`,
		Run:  devMappingRun(&f, factory),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.vscodeConfig, "vscode", "", "Path to VSCode keybindings.json file")
	cmd.Flags().StringVar(&f.xcodeConfig, "xcode", "", "Path to Xcode .idekeybindings file")
	cmd.Flags().StringVar(&f.outputFormat, "format", "table", "Output format: table, json")

	if err := cmd.MarkFlagRequired("vscode"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("xcode"); err != nil {
		panic(err)
	}

	return cmd
}

func devMappingRun(
	f *devMappingFlags,
	factory *keybindinglookup.LookupFactory,
) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		// 1. Load all keybindings from VSCode config
		allVSCodeKeybindings, err := loadAllVSCodeKeybindings(f.vscodeConfig)
		if err != nil {
			cmd.PrintErrf("Error loading VSCode keybindings: %v\n", err)
			os.Exit(1)
		}

		// 2. Create lookup instances for both editors
		vscodeLookup, err := factory.CreateLookup("vscode")
		if err != nil {
			cmd.PrintErrf("Error creating VSCode lookup: %v\n", err)
			os.Exit(1)
		}

		xcodeLookup, err := factory.CreateLookup("xcode")
		if err != nil {
			cmd.PrintErrf("Error creating Xcode lookup: %v\n", err)
			os.Exit(1)
		}

		// 3. For each keybinding, find matches in both editors
		mappings := findMappingsUsingLookup(
			allVSCodeKeybindings,
			f.vscodeConfig,
			f.xcodeConfig,
			vscodeLookup,
			xcodeLookup,
		)

		// 4. Output results
		if err := outputMappings(cmd, mappings, f.outputFormat); err != nil {
			cmd.PrintErrf("Error outputting results: %v\n", err)
			os.Exit(1)
		}
	}
}

// loadAllVSCodeKeybindings loads all keybindings from VSCode config file
// This returns the raw keybinding strings that we'll use to query both editors
func loadAllVSCodeKeybindings(configPath string) ([]string, error) {
	// Read the entire file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read VSCode config: %w", err)
	}

	// Remove comments (lines starting with //)
	lines := strings.Split(string(configData), "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanJSON := strings.Join(cleanLines, "\n")

	// Parse to temporary struct just to extract keys
	var tempBindings []struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal([]byte(cleanJSON), &tempBindings); err != nil {
		return nil, fmt.Errorf("failed to parse VSCode config: %w", err)
	}

	// Extract unique keybindings
	uniqueKeys := make(map[string]bool)
	var allKeys []string
	for _, binding := range tempBindings {
		if binding.Key != "" && !uniqueKeys[binding.Key] {
			uniqueKeys[binding.Key] = true
			allKeys = append(allKeys, binding.Key)
		}
	}

	return allKeys, nil
}

// findMappingsUsingLookup uses KeybindingLookup interface to find mappings
func findMappingsUsingLookup(
	allKeys []string,
	vscodeConfigPath string,
	xcodeConfigPath string,
	vscodeLookup keybindinglookup.KeybindingLookup,
	xcodeLookup keybindinglookup.KeybindingLookup,
) []KeybindingMapping {
	var mappings []KeybindingMapping

	for _, keyStr := range allKeys {
		// Parse the key string
		kb, err := keymap.ParseKeyBinding(keyStr, "+")
		if err != nil {
			continue // Skip invalid keybindings
		}

		// Look up in VSCode config
		vscodeFile, err := os.Open(vscodeConfigPath)
		if err != nil {
			continue
		}
		vscodeResults, err := vscodeLookup.Lookup(vscodeFile, kb)
		_ = vscodeFile.Close()
		if err != nil || len(vscodeResults) == 0 {
			continue // No VSCode matches
		}

		// Look up in Xcode config
		xcodeFile, err := os.Open(xcodeConfigPath)
		if err != nil {
			continue
		}
		xcodeResults, err := xcodeLookup.Lookup(xcodeFile, kb)
		_ = xcodeFile.Close()
		if err != nil || len(xcodeResults) == 0 {
			continue // No Xcode matches
		}

		// Create mappings for all combinations
		for _, vscodeResult := range vscodeResults {
			var vscodeBinding struct {
				Key     string `json:"key"`
				Command string `json:"command"`
				When    string `json:"when,omitempty"`
			}
			if err := json.Unmarshal([]byte(vscodeResult), &vscodeBinding); err != nil {
				continue
			}

			for _, xcodeResult := range xcodeResults {
				var xcodeBinding struct {
					Action           string `json:"Action"`
					KeyboardShortcut string `json:"Keyboard Shortcut"`
					Title            string `json:"Title"`
				}
				if err := json.Unmarshal([]byte(xcodeResult), &xcodeBinding); err != nil {
					continue
				}

				mapping := KeybindingMapping{
					Keybinding:  keyStr,
					VSCodeCmd:   vscodeBinding.Command,
					VSCodeWhen:  vscodeBinding.When,
					XcodeAction: xcodeBinding.Action,
					XcodeTitle:  xcodeBinding.Title,
				}
				mappings = append(mappings, mapping)
			}
		}
	}

	// Sort by keybinding for consistent output
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Keybinding < mappings[j].Keybinding
	})

	return mappings
}

func outputMappings(cmd *cobra.Command, mappings []KeybindingMapping, format string) error {
	if len(mappings) == 0 {
		cmd.Println("No matching keybindings found between the configurations.")
		return nil
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(mappings)

	case "table":
		cmd.Printf("Found %d matching keybindings:\n\n", len(mappings))

		// Calculate column widths
		maxKeybinding := len("Keybinding")
		maxVSCode := len("VSCode Command")
		maxXcode := len("Xcode Action")

		for _, m := range mappings {
			if len(m.Keybinding) > maxKeybinding {
				maxKeybinding = len(m.Keybinding)
			}
			if len(m.VSCodeCmd) > maxVSCode {
				maxVSCode = len(m.VSCodeCmd)
			}
			if len(m.XcodeAction) > maxXcode {
				maxXcode = len(m.XcodeAction)
			}
		}

		// Print header
		tableFormat := fmt.Sprintf("%%-%ds | %%-%ds | %%-%ds\n", maxKeybinding, maxVSCode, maxXcode)
		cmd.Printf(tableFormat, "Keybinding", "VSCode Command", "Xcode Action")
		const separatorPadding = 6
		cmd.Printf("%s\n", strings.Repeat("-", maxKeybinding+maxVSCode+maxXcode+separatorPadding))

		// Print mappings
		for _, m := range mappings {
			cmd.Printf(tableFormat, m.Keybinding, m.VSCodeCmd, m.XcodeAction)
		}

		return nil

	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}
