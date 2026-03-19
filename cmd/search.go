package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Lookup search endpoints (apps, geolocations)",
}
