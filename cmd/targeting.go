package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(targetingCmd)
}

var targetingCmd = &cobra.Command{
	Use:   "targeting",
	Short: "Manage ad-group targeting dimensions",
}
