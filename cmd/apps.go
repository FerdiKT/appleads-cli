package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(appsCmd)
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "App-level Apple Ads endpoints",
}
