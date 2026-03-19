package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(campaignsCmd)
}

var campaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Manage Apple Ads campaigns",
}
