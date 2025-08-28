/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
)

const validZedActionsPath = "onekeymap/chore/zed-valid-action.json"

// zedActionValidateCmd represents the zedActionValidate command
var zedActionValidateCmd = &cobra.Command{
	Use:   "zedActionValidate",
	Short: "Validates that all Zed actions in the mappings are valid.",
	Long: `This command checks all action mapping configuration files to ensure that every
defined 'zed.action' corresponds to a valid, known action from the generated list.

It reads the list of valid actions from 'onekeymap/chore/zed-valid-action.json'.
If an invalid action is found, it will report the error and exit with a non-zero status code.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateZedActions(); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ All Zed actions in mappings are valid.")
	},
}

func validateZedActions() error {
	// 1. Read and parse the valid actions JSON file
	jsonFile, err := os.Open(validZedActionsPath)
	if err != nil {
		return fmt.Errorf("could not open valid zed actions file: %w", err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, _ := io.ReadAll(jsonFile)
	var validActions []string
	if err := json.Unmarshal(byteValue, &validActions); err != nil {
		return fmt.Errorf("could not unmarshal valid actions json: %w", err)
	}

	// Convert slice to a map for efficient lookup
	validActionsSet := make(map[string]struct{}, len(validActions))
	for _, action := range validActions {
		validActionsSet[action] = struct{}{}
	}

	// 2. Load all action mappings
	config, err := mappings.NewMappingConfig()
	if err != nil {
		return fmt.Errorf("could not load mapping config: %w", err)
	}

	// 3. Iterate and validate each mapping (support multiple Zed configs per action)
	var invalidActionsFound bool
	for _, mapping := range config.Mappings {
		for _, zconf := range mapping.Zed {
			if zconf.Action == "" {
				continue // Skip if no zed action is defined for this entry
			}
			if _, ok := validActionsSet[zconf.Action]; !ok {
				fmt.Fprintf(os.Stderr, "Error: Invalid Zed action '%s' (context: '%s') found in mapping '%s'\n", zconf.Action, zconf.Context, mapping.ID)
				invalidActionsFound = true
			}
		}
	}

	if invalidActionsFound {
		return fmt.Errorf("invalid zed actions were found")
	}

	return nil
}

func init() {
	devCmd.AddCommand(zedActionValidateCmd)
}
