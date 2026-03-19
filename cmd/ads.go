package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(adsCmd)
}

var adsCmd = &cobra.Command{
	Use:   "ads",
	Short: "Manage Apple Ads creatives in ad groups",
}
