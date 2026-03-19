package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type reportCallFlags struct {
	OrgID      int64
	CampaignID int64
	Body       string
	BodyFile   string
}

var reportsCampaignsFlags reportCallFlags
var reportsAdGroupsFlags reportCallFlags
var reportsKeywordsFlags reportCallFlags
var reportsSearchTermsFlags reportCallFlags
var reportsAdsFlags reportCallFlags
var reportsImpressionShareFlags reportCallFlags

func init() {
	reportsCmd.AddCommand(reportsCampaignsCmd)
	reportsCmd.AddCommand(reportsAdGroupsCmd)
	reportsCmd.AddCommand(reportsKeywordsCmd)
	reportsCmd.AddCommand(reportsSearchTermsCmd)
	reportsCmd.AddCommand(reportsAdsCmd)
	reportsCmd.AddCommand(reportsImpressionShareCmd)

	addReportFlags(reportsCampaignsCmd, &reportsCampaignsFlags, false)
	addReportFlags(reportsAdGroupsCmd, &reportsAdGroupsFlags, true)
	addReportFlags(reportsKeywordsCmd, &reportsKeywordsFlags, true)
	addReportFlags(reportsSearchTermsCmd, &reportsSearchTermsFlags, true)
	addReportFlags(reportsAdsCmd, &reportsAdsFlags, true)
	addReportFlags(reportsImpressionShareCmd, &reportsImpressionShareFlags, true)
}

func addReportFlags(cmd *cobra.Command, flags *reportCallFlags, campaignRequired bool) {
	cmd.Flags().Int64Var(&flags.OrgID, "org-id", 0, "Organization ID override")
	cmd.Flags().Int64Var(&flags.CampaignID, "campaign-id", 0, "Campaign ID")
	cmd.Flags().StringVar(&flags.Body, "body", "", "Inline JSON report request")
	cmd.Flags().StringVar(&flags.BodyFile, "body-file", "", "Path to JSON report request file")
	if campaignRequired {
		_ = cmd.MarkFlagRequired("campaign-id")
	}
}

var reportsCampaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Run campaign-level report",
	Example: "  appleads reports campaigns --body-file ./payloads/report-campaigns.json\n" +
		"  appleads reports campaigns --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\",\"granularity\":\"DAILY\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportCall("/reports/campaigns", reportsCampaignsFlags, false)
	},
}

var reportsAdGroupsCmd = &cobra.Command{
	Use:   "adgroups",
	Short: "Run ad-group-level report for a campaign",
	Example: "  appleads reports adgroups --campaign-id 123456 --body-file ./payloads/report-adgroups.json\n" +
		"  appleads reports adgroups --campaign-id 123456 --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportCall(fmt.Sprintf("/reports/campaigns/%d/adgroups", reportsAdGroupsFlags.CampaignID), reportsAdGroupsFlags, true)
	},
}

var reportsKeywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Run keyword-level report for a campaign",
	Example: "  appleads reports keywords --campaign-id 123456 --body-file ./payloads/report-keywords.json\n" +
		"  appleads reports keywords --campaign-id 123456 --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportCall(fmt.Sprintf("/reports/campaigns/%d/keywords", reportsKeywordsFlags.CampaignID), reportsKeywordsFlags, true)
	},
}

var reportsSearchTermsCmd = &cobra.Command{
	Use:   "searchterms",
	Short: "Run search-term-level report for a campaign",
	Example: "  appleads reports searchterms --campaign-id 123456 --body-file ./payloads/report-searchterms.json\n" +
		"  appleads reports searchterms --campaign-id 123456 --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportCall(fmt.Sprintf("/reports/campaigns/%d/searchterms", reportsSearchTermsFlags.CampaignID), reportsSearchTermsFlags, true)
	},
}

var reportsAdsCmd = &cobra.Command{
	Use:   "ads",
	Short: "Run ad-level report for a campaign",
	Example: "  appleads reports ads --campaign-id 123456 --body-file ./payloads/report-ads.json\n" +
		"  appleads reports ads --campaign-id 123456 --body '{\"startTime\":\"2026-03-01\",\"endTime\":\"2026-03-17\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportCall(fmt.Sprintf("/reports/campaigns/%d/ads", reportsAdsFlags.CampaignID), reportsAdsFlags, true)
	},
}

var reportsImpressionShareCmd = &cobra.Command{
	Use:   "impressionshare",
	Short: "Run impression-share report for a campaign",
	Example: "  appleads reports impressionshare --campaign-id 123456 --body-file ./payloads/report-impressionshare.json\n" +
		"  appleads reports template impressionshare --campaign-id 123456 --run",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportsImpressionShareFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		payload, err := readJSONPayload(reportsImpressionShareFlags.Body, reportsImpressionShareFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), reportsImpressionShareFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{
			fmt.Sprintf("/reports/campaigns/%d/impressionshare", reportsImpressionShareFlags.CampaignID),
			fmt.Sprintf("/reports/campaigns/%d/impressionShare", reportsImpressionShareFlags.CampaignID),
			fmt.Sprintf("/reports/campaigns/%d/impression-share", reportsImpressionShareFlags.CampaignID),
			fmt.Sprintf("/reports/campaigns/%d/impression_share", reportsImpressionShareFlags.CampaignID),
		}
		if err := callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, paths, nil, payload); err != nil {
			return fmt.Errorf("impression share report endpoint may be unavailable for this account/api version: %w", err)
		}
		return nil
	},
}

func runReportCall(path string, flags reportCallFlags, campaignRequired bool) error {
	if campaignRequired && flags.CampaignID <= 0 {
		return fmt.Errorf("--campaign-id must be > 0")
	}

	payload, err := readJSONPayload(flags.Body, flags.BodyFile, false)
	if err != nil {
		return err
	}

	client, _, _, err := authedClient(context.Background(), flags.OrgID, true)
	if err != nil {
		return err
	}

	return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
}
