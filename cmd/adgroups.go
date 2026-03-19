package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(adGroupsCmd)
}

var adGroupsCmd = &cobra.Command{
	Use:   "adgroups",
	Short: "Manage Apple Ads ad groups",
}
