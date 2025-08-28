package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OneKeymap CLI\n")
		fmt.Printf("Version:    %s\n", version)
		fmt.Printf("Built:      %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Printf("Go Version: %s\n", "go1.23.5") // Could be dynamic if needed
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
