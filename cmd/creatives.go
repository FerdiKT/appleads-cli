package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(creativesCmd)
}

var creativesCmd = &cobra.Command{
	Use:   "creatives",
	Short: "Manage creatives",
}
