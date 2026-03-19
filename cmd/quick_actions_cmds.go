package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var campaignsEnableFlags struct {
	OrgID  int64
	DryRun bool
	Yes    bool
}

var campaignsPauseFlags struct {
	OrgID  int64
	DryRun bool
	Yes    bool
}

var adgroupsEnableFlags struct {
	OrgID      int64
	CampaignID int64
	DryRun     bool
	Yes        bool
}

var adgroupsPauseFlags struct {
	OrgID      int64
	CampaignID int64
	DryRun     bool
	Yes        bool
}

var adsEnableFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

var adsPauseFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

var kwTargetEnableFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

var kwTargetPauseFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	DryRun     bool
	Yes        bool
}

func init() {
	campaignsCmd.AddCommand(campaignsEnableCmd)
	campaignsCmd.AddCommand(campaignsPauseCmd)
	adGroupsCmd.AddCommand(adgroupsEnableCmd)
	adGroupsCmd.AddCommand(adgroupsPauseCmd)
	adsCmd.AddCommand(adsEnableCmd)
	adsCmd.AddCommand(adsPauseCmd)
	keywordsTargetingCmd.AddCommand(kwTargetEnableCmd)
	keywordsTargetingCmd.AddCommand(kwTargetPauseCmd)

	addStatusMutationFlags(campaignsEnableCmd, &campaignsEnableFlags.OrgID, &campaignsEnableFlags.DryRun, &campaignsEnableFlags.Yes)
	addStatusMutationFlags(campaignsPauseCmd, &campaignsPauseFlags.OrgID, &campaignsPauseFlags.DryRun, &campaignsPauseFlags.Yes)

	addStatusMutationFlags(adgroupsEnableCmd, &adgroupsEnableFlags.OrgID, &adgroupsEnableFlags.DryRun, &adgroupsEnableFlags.Yes)
	adgroupsEnableCmd.Flags().Int64Var(&adgroupsEnableFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	_ = adgroupsEnableCmd.MarkFlagRequired("campaign-id")
	addStatusMutationFlags(adgroupsPauseCmd, &adgroupsPauseFlags.OrgID, &adgroupsPauseFlags.DryRun, &adgroupsPauseFlags.Yes)
	adgroupsPauseCmd.Flags().Int64Var(&adgroupsPauseFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	_ = adgroupsPauseCmd.MarkFlagRequired("campaign-id")

	addStatusMutationFlags(adsEnableCmd, &adsEnableFlags.OrgID, &adsEnableFlags.DryRun, &adsEnableFlags.Yes)
	adsEnableCmd.Flags().Int64Var(&adsEnableFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsEnableCmd.Flags().Int64Var(&adsEnableFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	_ = adsEnableCmd.MarkFlagRequired("campaign-id")
	_ = adsEnableCmd.MarkFlagRequired("adgroup-id")
	addStatusMutationFlags(adsPauseCmd, &adsPauseFlags.OrgID, &adsPauseFlags.DryRun, &adsPauseFlags.Yes)
	adsPauseCmd.Flags().Int64Var(&adsPauseFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsPauseCmd.Flags().Int64Var(&adsPauseFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	_ = adsPauseCmd.MarkFlagRequired("campaign-id")
	_ = adsPauseCmd.MarkFlagRequired("adgroup-id")

	addStatusMutationFlags(kwTargetEnableCmd, &kwTargetEnableFlags.OrgID, &kwTargetEnableFlags.DryRun, &kwTargetEnableFlags.Yes)
	kwTargetEnableCmd.Flags().Int64Var(&kwTargetEnableFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetEnableCmd.Flags().Int64Var(&kwTargetEnableFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	_ = kwTargetEnableCmd.MarkFlagRequired("campaign-id")
	_ = kwTargetEnableCmd.MarkFlagRequired("adgroup-id")
	addStatusMutationFlags(kwTargetPauseCmd, &kwTargetPauseFlags.OrgID, &kwTargetPauseFlags.DryRun, &kwTargetPauseFlags.Yes)
	kwTargetPauseCmd.Flags().Int64Var(&kwTargetPauseFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetPauseCmd.Flags().Int64Var(&kwTargetPauseFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	_ = kwTargetPauseCmd.MarkFlagRequired("campaign-id")
	_ = kwTargetPauseCmd.MarkFlagRequired("adgroup-id")
}

func addStatusMutationFlags(cmd *cobra.Command, orgID *int64, dryRun, yes *bool) {
	cmd.Flags().Int64Var(orgID, "org-id", 0, "Organization ID override")
	cmd.Flags().BoolVar(dryRun, "dry-run", false, "Print payload without sending update request")
	cmd.Flags().BoolVar(yes, "yes", false, "Skip interactive confirmation")
}

var campaignsEnableCmd = &cobra.Command{
	Use:   "enable <campaign-id>",
	Short: "Quick action: set campaign status to ENABLED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		campaignID, err := parsePositiveInt64("campaign-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d", campaignID)
		return runSimpleStatusMutation(campaignsEnableFlags.OrgID, path, map[string]any{"status": "ENABLED"}, campaignsEnableFlags.DryRun, campaignsEnableFlags.Yes, "campaign enable")
	},
}

var campaignsPauseCmd = &cobra.Command{
	Use:   "pause <campaign-id>",
	Short: "Quick action: set campaign status to PAUSED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		campaignID, err := parsePositiveInt64("campaign-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d", campaignID)
		return runSimpleStatusMutation(campaignsPauseFlags.OrgID, path, map[string]any{"status": "PAUSED"}, campaignsPauseFlags.DryRun, campaignsPauseFlags.Yes, "campaign pause")
	},
}

var adgroupsEnableCmd = &cobra.Command{
	Use:   "enable <adgroup-id>",
	Short: "Quick action: set ad group status to ENABLED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if adgroupsEnableFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		adGroupID, err := parsePositiveInt64("adgroup-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d", adgroupsEnableFlags.CampaignID, adGroupID)
		return runSimpleStatusMutation(adgroupsEnableFlags.OrgID, path, map[string]any{"status": "ENABLED"}, adgroupsEnableFlags.DryRun, adgroupsEnableFlags.Yes, "adgroup enable")
	},
}

var adgroupsPauseCmd = &cobra.Command{
	Use:   "pause <adgroup-id>",
	Short: "Quick action: set ad group status to PAUSED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if adgroupsPauseFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		adGroupID, err := parsePositiveInt64("adgroup-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d", adgroupsPauseFlags.CampaignID, adGroupID)
		return runSimpleStatusMutation(adgroupsPauseFlags.OrgID, path, map[string]any{"status": "PAUSED"}, adgroupsPauseFlags.DryRun, adgroupsPauseFlags.Yes, "adgroup pause")
	},
}

var adsEnableCmd = &cobra.Command{
	Use:   "enable <ad-id>",
	Short: "Quick action: set ad status to ENABLED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if adsEnableFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if adsEnableFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		adID, err := parsePositiveInt64("ad-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/ads/%d", adsEnableFlags.CampaignID, adsEnableFlags.AdGroupID, adID)
		return runSimpleStatusMutation(adsEnableFlags.OrgID, path, map[string]any{"status": "ENABLED"}, adsEnableFlags.DryRun, adsEnableFlags.Yes, "ad enable")
	},
}

var adsPauseCmd = &cobra.Command{
	Use:   "pause <ad-id>",
	Short: "Quick action: set ad status to PAUSED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if adsPauseFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if adsPauseFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		adID, err := parsePositiveInt64("ad-id", args[0])
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/ads/%d", adsPauseFlags.CampaignID, adsPauseFlags.AdGroupID, adID)
		return runSimpleStatusMutation(adsPauseFlags.OrgID, path, map[string]any{"status": "PAUSED"}, adsPauseFlags.DryRun, adsPauseFlags.Yes, "ad pause")
	},
}

var kwTargetEnableCmd = &cobra.Command{
	Use:   "enable <keyword-id>",
	Short: "Quick action: set targeting keyword status to ACTIVE",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetEnableFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwTargetEnableFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		keywordID, err := parsePositiveInt64("keyword-id", args[0])
		if err != nil {
			return err
		}
		payload := []any{
			map[string]any{
				"id":     keywordID,
				"status": "ACTIVE",
			},
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/bulk", kwTargetEnableFlags.CampaignID, kwTargetEnableFlags.AdGroupID)
		return runSimpleStatusMutation(kwTargetEnableFlags.OrgID, path, payload, kwTargetEnableFlags.DryRun, kwTargetEnableFlags.Yes, "targeting keyword enable")
	},
}

var kwTargetPauseCmd = &cobra.Command{
	Use:   "pause <keyword-id>",
	Short: "Quick action: set targeting keyword status to PAUSED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetPauseFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwTargetPauseFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		keywordID, err := parsePositiveInt64("keyword-id", args[0])
		if err != nil {
			return err
		}
		payload := []any{
			map[string]any{
				"id":     keywordID,
				"status": "PAUSED",
			},
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/bulk", kwTargetPauseFlags.CampaignID, kwTargetPauseFlags.AdGroupID)
		return runSimpleStatusMutation(kwTargetPauseFlags.OrgID, path, payload, kwTargetPauseFlags.DryRun, kwTargetPauseFlags.Yes, "targeting keyword pause")
	},
}

func runSimpleStatusMutation(orgID int64, path string, payload any, dryRun, yes bool, title string) error {
	if dryRun {
		return printJSON(payload)
	}
	if err := confirmJSONPayload(title, payload, yes); err != nil {
		return err
	}
	client, _, _, err := authedClient(context.Background(), orgID, true)
	if err != nil {
		return err
	}
	return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
}
