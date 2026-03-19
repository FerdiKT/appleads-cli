package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(reportsCmd)
}

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Run Apple Ads reporting endpoints",
}
