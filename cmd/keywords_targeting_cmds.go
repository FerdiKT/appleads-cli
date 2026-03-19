package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var kwTargetListFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Offset     int
	Limit      int
	All        bool
}

var kwTargetGetFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
}

var kwTargetCreateFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

var kwTargetUpdateFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

var kwTargetDeleteFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

var kwTargetFindFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

func init() {
	keywordsTargetingCmd.AddCommand(kwTargetListCmd)
	keywordsTargetingCmd.AddCommand(kwTargetGetCmd)
	keywordsTargetingCmd.AddCommand(kwTargetCreateCmd)
	keywordsTargetingCmd.AddCommand(kwTargetUpdateCmd)
	keywordsTargetingCmd.AddCommand(kwTargetDeleteCmd)
	keywordsTargetingCmd.AddCommand(kwTargetFindCmd)

	kwTargetListCmd.Flags().Int64Var(&kwTargetListFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetListCmd.Flags().Int64Var(&kwTargetListFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetListCmd.Flags().Int64Var(&kwTargetListFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwTargetListCmd.Flags().IntVar(&kwTargetListFlags.Offset, "offset", 0, "Pagination offset")
	kwTargetListCmd.Flags().IntVar(&kwTargetListFlags.Limit, "limit", 20, "Pagination limit")
	kwTargetListCmd.Flags().BoolVar(&kwTargetListFlags.All, "all", false, "Fetch all pages")

	kwTargetGetCmd.Flags().Int64Var(&kwTargetGetFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetGetCmd.Flags().Int64Var(&kwTargetGetFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetGetCmd.Flags().Int64Var(&kwTargetGetFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")

	kwTargetCreateCmd.Flags().Int64Var(&kwTargetCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetCreateCmd.Flags().Int64Var(&kwTargetCreateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetCreateCmd.Flags().Int64Var(&kwTargetCreateFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwTargetCreateCmd.Flags().StringVar(&kwTargetCreateFlags.Body, "body", "", "Inline JSON body (usually array)")
	kwTargetCreateCmd.Flags().StringVar(&kwTargetCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")
	_ = kwTargetCreateCmd.MarkFlagRequired("campaign-id")
	_ = kwTargetCreateCmd.MarkFlagRequired("adgroup-id")

	kwTargetUpdateCmd.Flags().Int64Var(&kwTargetUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetUpdateCmd.Flags().Int64Var(&kwTargetUpdateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetUpdateCmd.Flags().Int64Var(&kwTargetUpdateFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwTargetUpdateCmd.Flags().StringVar(&kwTargetUpdateFlags.Body, "body", "", "Inline JSON body (usually array)")
	kwTargetUpdateCmd.Flags().StringVar(&kwTargetUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")
	_ = kwTargetUpdateCmd.MarkFlagRequired("campaign-id")
	_ = kwTargetUpdateCmd.MarkFlagRequired("adgroup-id")

	kwTargetDeleteCmd.Flags().Int64Var(&kwTargetDeleteFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetDeleteCmd.Flags().Int64Var(&kwTargetDeleteFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwTargetDeleteCmd.Flags().Int64Var(&kwTargetDeleteFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	kwTargetDeleteCmd.Flags().StringVar(&kwTargetDeleteFlags.Body, "body", "", "Inline JSON body (usually array of IDs/objects)")
	kwTargetDeleteCmd.Flags().StringVar(&kwTargetDeleteFlags.BodyFile, "body-file", "", "Path to JSON body file")
	_ = kwTargetDeleteCmd.MarkFlagRequired("campaign-id")
	_ = kwTargetDeleteCmd.MarkFlagRequired("adgroup-id")

	kwTargetFindCmd.Flags().Int64Var(&kwTargetFindFlags.OrgID, "org-id", 0, "Organization ID override")
	kwTargetFindCmd.Flags().Int64Var(&kwTargetFindFlags.CampaignID, "campaign-id", 0, "Campaign ID (optional)")
	kwTargetFindCmd.Flags().Int64Var(&kwTargetFindFlags.AdGroupID, "adgroup-id", 0, "Ad group ID (optional)")
	kwTargetFindCmd.Flags().StringVar(&kwTargetFindFlags.Body, "body", "", "Inline JSON selector body")
	kwTargetFindCmd.Flags().StringVar(&kwTargetFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
}

var kwTargetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List targeting keywords (org/campaign/ad-group scope based on flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetListFlags.AdGroupID > 0 && kwTargetListFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0 when --adgroup-id is provided")
		}
		if kwTargetListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if kwTargetListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}

		client, _, _, err := authedClient(context.Background(), kwTargetListFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(kwTargetListFlags.Offset))
		q.Set("limit", strconv.Itoa(kwTargetListFlags.Limit))
		if kwTargetListFlags.CampaignID > 0 && kwTargetListFlags.AdGroupID > 0 {
			path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords", kwTargetListFlags.CampaignID, kwTargetListFlags.AdGroupID)
			return callListEndpoint(context.Background(), client, path, q, kwTargetListFlags.Offset, kwTargetListFlags.Limit, kwTargetListFlags.All)
		}
		if kwTargetListFlags.CampaignID > 0 {
			err = callListEndpointWithFallback(context.Background(), client, []string{
				fmt.Sprintf("/campaigns/%d/targetingkeywords", kwTargetListFlags.CampaignID),
				"/targetingkeywords",
			}, q, kwTargetListFlags.Offset, kwTargetListFlags.Limit, kwTargetListFlags.All)
			return withNotFoundHint(err, "campaign-level targeting keywords endpoint unavailable for this account, try with --adgroup-id")
		}
		err = callListEndpoint(context.Background(), client, "/targetingkeywords", q, kwTargetListFlags.Offset, kwTargetListFlags.Limit, kwTargetListFlags.All)
		return withNotFoundHint(err, "org-level targeting keywords endpoint unavailable for this account, try with --campaign-id and --adgroup-id")
	},
}

var kwTargetGetCmd = &cobra.Command{
	Use:   "get <keyword-id>",
	Short: "Get targeting keyword by ID (scoped path is used when flags are provided)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetGetFlags.AdGroupID > 0 && kwTargetGetFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0 when --adgroup-id is provided")
		}
		keywordID, err := parsePositiveInt64("keyword-id", args[0])
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), kwTargetGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{}
		if kwTargetGetFlags.CampaignID > 0 && kwTargetGetFlags.AdGroupID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/%d", kwTargetGetFlags.CampaignID, kwTargetGetFlags.AdGroupID, keywordID))
		}
		if kwTargetGetFlags.CampaignID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/targetingkeywords/%d", kwTargetGetFlags.CampaignID, keywordID))
		}
		paths = append(paths, fmt.Sprintf("/targetingkeywords/%d", keywordID))
		if len(paths) == 1 {
			return callAPIAndPrint(context.Background(), client, http.MethodGet, paths[0], nil, nil)
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodGet, paths, nil, nil)
	},
}

var kwTargetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create targeting keywords (bulk) in an ad group",
	Example: "  appleads keywords targeting create --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-target-create.json\n" +
		"  appleads keywords targeting create --campaign-id 123456 --adgroup-id 987654 --body '[{\"text\":\"my keyword\",\"matchType\":\"EXACT\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetCreateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwTargetCreateFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwTargetCreateFlags.Body, kwTargetCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwTargetCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/bulk", kwTargetCreateFlags.CampaignID, kwTargetCreateFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}

var kwTargetUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update targeting keywords (bulk) in an ad group",
	Example: "  appleads keywords targeting update --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-target-update.json\n" +
		"  appleads keywords targeting update --campaign-id 123456 --adgroup-id 987654 --body '[{\"id\":111111,\"status\":\"PAUSED\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetUpdateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwTargetUpdateFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwTargetUpdateFlags.Body, kwTargetUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwTargetUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/bulk", kwTargetUpdateFlags.CampaignID, kwTargetUpdateFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
	},
}

var kwTargetDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete targeting keywords (bulk) in an ad group",
	Example: "  appleads keywords targeting delete --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-target-delete.json\n" +
		"  appleads keywords targeting delete --campaign-id 123456 --adgroup-id 987654 --body '[{\"id\":111111}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwTargetDeleteFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwTargetDeleteFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwTargetDeleteFlags.Body, kwTargetDeleteFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwTargetDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/bulk", kwTargetDeleteFlags.CampaignID, kwTargetDeleteFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodDelete, path, nil, payload)
	},
}

var kwTargetFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find targeting keywords by selector (org, campaign, or ad-group level)",
	Example: "  appleads keywords targeting find --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-target-find.json\n" +
		"  appleads keywords targeting find --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(kwTargetFindFlags.Body, kwTargetFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwTargetFindFlags.OrgID, true)
		if err != nil {
			return err
		}

		path := "/targetingkeywords/find"
		if kwTargetFindFlags.CampaignID > 0 && kwTargetFindFlags.AdGroupID > 0 {
			path = fmt.Sprintf("/campaigns/%d/adgroups/%d/targetingkeywords/find", kwTargetFindFlags.CampaignID, kwTargetFindFlags.AdGroupID)
		} else if kwTargetFindFlags.CampaignID > 0 {
			path = fmt.Sprintf("/campaigns/%d/targetingkeywords/find", kwTargetFindFlags.CampaignID)
		}

		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}
