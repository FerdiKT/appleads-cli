package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var kwRecListFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Offset     int
	Limit      int
	All        bool
}

var kwRecFindFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

func init() {
	keywordsRecommendationsCmd.AddCommand(kwRecListCmd)
	keywordsRecommendationsCmd.AddCommand(kwRecFindCmd)

	kwRecListCmd.Flags().Int64Var(&kwRecListFlags.OrgID, "org-id", 0, "Organization ID override")
	kwRecListCmd.Flags().Int64Var(&kwRecListFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwRecListCmd.Flags().Int64Var(&kwRecListFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwRecListCmd.Flags().IntVar(&kwRecListFlags.Offset, "offset", 0, "Pagination offset")
	kwRecListCmd.Flags().IntVar(&kwRecListFlags.Limit, "limit", 20, "Pagination limit")
	kwRecListCmd.Flags().BoolVar(&kwRecListFlags.All, "all", false, "Fetch all pages")
	_ = kwRecListCmd.MarkFlagRequired("campaign-id")
	_ = kwRecListCmd.MarkFlagRequired("adgroup-id")

	kwRecFindCmd.Flags().Int64Var(&kwRecFindFlags.OrgID, "org-id", 0, "Organization ID override")
	kwRecFindCmd.Flags().Int64Var(&kwRecFindFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwRecFindCmd.Flags().Int64Var(&kwRecFindFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwRecFindCmd.Flags().StringVar(&kwRecFindFlags.Body, "body", "", "Inline JSON selector body")
	kwRecFindCmd.Flags().StringVar(&kwRecFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
	_ = kwRecFindCmd.MarkFlagRequired("campaign-id")
	_ = kwRecFindCmd.MarkFlagRequired("adgroup-id")
}

var kwRecListCmd = &cobra.Command{
	Use:   "list",
	Short: "List keyword recommendations for an ad group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwRecListFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwRecListFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		if kwRecListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if kwRecListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), kwRecListFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(kwRecListFlags.Offset))
		q.Set("limit", strconv.Itoa(kwRecListFlags.Limit))
		paths := []string{
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/recommendations", kwRecListFlags.CampaignID, kwRecListFlags.AdGroupID),
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/suggestions", kwRecListFlags.CampaignID, kwRecListFlags.AdGroupID),
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/recommended", kwRecListFlags.CampaignID, kwRecListFlags.AdGroupID),
		}
		return callListEndpointWithFallback(context.Background(), client, paths, q, kwRecListFlags.Offset, kwRecListFlags.Limit, kwRecListFlags.All)
	},
}

var kwRecFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find keyword recommendations with selector payload",
	Example: "  appleads keywords recommendations find --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-recommendations-find.json\n" +
		"  appleads keywords recommendations find --campaign-id 123456 --adgroup-id 987654 --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwRecFindFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwRecFindFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwRecFindFlags.Body, kwRecFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwRecFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/recommendations/find", kwRecFindFlags.CampaignID, kwRecFindFlags.AdGroupID),
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/suggestions/find", kwRecFindFlags.CampaignID, kwRecFindFlags.AdGroupID),
			fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/recommended/find", kwRecFindFlags.CampaignID, kwRecFindFlags.AdGroupID),
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, paths, nil, payload)
	},
}
