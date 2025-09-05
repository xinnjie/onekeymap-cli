package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
)

const validZedActionsPath = "onekeymap/onekeymap-cli/chore/zed-valid-action.json"

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Runs diagnostic checks on mapping configurations.",
	Long: `The doctor command runs a series of diagnostic checks on the action mapping configurations.

It currently performs two main checks:
1. Description Check: Verifies that all mappings have a 'description' and 'short_description'.
2. Zed Action Validation: Ensures that all 'zed.action' entries correspond to valid, known actions.

This command is essential for maintaining the quality, consistency, and correctness of the keymap configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running mapping configuration diagnostics...")
		var hasErrors bool

		if err := checkDescriptions(); err != nil {
			fmt.Fprintf(os.Stderr, "\n[FAIL] Description check failed: %v\n", err)
			hasErrors = true
		}

		if err := validateZedActions(); err != nil {
			fmt.Fprintf(os.Stderr, "\n[FAIL] Zed action validation failed: %v\n", err)
			hasErrors = true
		}

		if hasErrors {
			fmt.Fprintf(os.Stderr, "\nDoctor checks completed with errors.\n")
			os.Exit(1)
		} else {
			fmt.Println("\n✅ All doctor checks passed successfully.")
		}
	},
}

func checkDescriptions() error {
	config, err := mappings.NewMappingConfig()
	if err != nil {
		return fmt.Errorf("could not load mapping config: %w", err)
	}

	var foundMissing bool
	fmt.Println("\nRunning Description and Short Description check...")

	for id, mapping := range config.Mappings {
		if mapping.Description == "" {
			fmt.Printf("  - [Missing Description] Action ID: %s\n", id)
			foundMissing = true
		}
		if mapping.ShortDescription == "" {
			fmt.Printf("  - [Missing Short Description] Action ID: %s\n", id)
			foundMissing = true
		}
	}

	if !foundMissing {
		fmt.Println("  => ✅ All mappings have descriptions and short descriptions.")
	} else {
		return fmt.Errorf("found mappings with missing descriptions")
	}

	return nil
}

func validateZedActions() error {
	fmt.Println("\nRunning Zed Action Validation check...")
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
				fmt.Fprintf(os.Stderr, "  - [Invalid Zed Action] Action: '%s', Context: '%s', Mapping ID: '%s'\n", zconf.Action, zconf.Context, mapping.ID)
				invalidActionsFound = true
			}
		}
	}

	if invalidActionsFound {
		return fmt.Errorf("invalid zed actions were found")
	}

	fmt.Println("  => ✅ All Zed actions in mappings are valid.")
	return nil
}

func init() {
	devCmd.AddCommand(doctorCmd)
}
