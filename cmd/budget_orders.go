package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(budgetOrdersCmd)
}

var budgetOrdersCmd = &cobra.Command{
	Use:   "budget-orders",
	Short: "Manage budget orders",
}
