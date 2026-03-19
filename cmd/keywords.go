package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(keywordsCmd)
	keywordsCmd.AddCommand(keywordsTargetingCmd)
	keywordsCmd.AddCommand(keywordsCampaignNegativeCmd)
	keywordsCmd.AddCommand(keywordsAdGroupNegativeCmd)
	keywordsCmd.AddCommand(keywordsRecommendationsCmd)
}

var keywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Manage Apple Ads keywords",
}

var keywordsTargetingCmd = &cobra.Command{
	Use:   "targeting",
	Short: "Manage targeting keywords",
}

var keywordsCampaignNegativeCmd = &cobra.Command{
	Use:   "campaign-negative",
	Short: "Manage campaign-level negative keywords",
}

var keywordsAdGroupNegativeCmd = &cobra.Command{
	Use:   "adgroup-negative",
	Short: "Manage ad-group-level negative keywords",
}

var keywordsRecommendationsCmd = &cobra.Command{
	Use:   "recommendations",
	Short: "Keyword recommendation endpoints",
}
