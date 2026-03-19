package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

type campaignNegativeFlags struct {
	OrgID      int64
	CampaignID int64
	Offset     int
	Limit      int
	All        bool
	Body       string
	BodyFile   string
}

type adGroupNegativeFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Offset     int
	Limit      int
	All        bool
	Body       string
	BodyFile   string
}

var kwCampNegListFlags campaignNegativeFlags
var kwCampNegGetFlags campaignNegativeFlags
var kwCampNegCreateFlags campaignNegativeFlags
var kwCampNegUpdateFlags campaignNegativeFlags
var kwCampNegDeleteFlags campaignNegativeFlags
var kwCampNegFindFlags campaignNegativeFlags

var kwCampNegEnableFlags struct {
	OrgID      int64
	CampaignID int64
	DryRun     bool
	Yes        bool
}

var kwCampNegPauseFlags struct {
	OrgID      int64
	CampaignID int64
	DryRun     bool
	Yes        bool
}

var kwAdgNegListFlags adGroupNegativeFlags
var kwAdgNegGetFlags adGroupNegativeFlags
var kwAdgNegCreateFlags adGroupNegativeFlags
var kwAdgNegUpdateFlags adGroupNegativeFlags
var kwAdgNegDeleteFlags adGroupNegativeFlags
var kwAdgNegFindFlags adGroupNegativeFlags

func init() {
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegListCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegGetCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegCreateCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegUpdateCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegDeleteCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegFindCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegEnableCmd)
	keywordsCampaignNegativeCmd.AddCommand(kwCampNegPauseCmd)

	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegListCmd)
	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegGetCmd)
	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegCreateCmd)
	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegUpdateCmd)
	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegDeleteCmd)
	keywordsAdGroupNegativeCmd.AddCommand(kwAdgNegFindCmd)

	addCampaignNegativeFlags(kwCampNegListCmd, &kwCampNegListFlags, false, true)
	addCampaignNegativeFlags(kwCampNegGetCmd, &kwCampNegGetFlags, false, false)
	addCampaignNegativeFlags(kwCampNegCreateCmd, &kwCampNegCreateFlags, true, false)
	addCampaignNegativeFlags(kwCampNegUpdateCmd, &kwCampNegUpdateFlags, true, false)
	addCampaignNegativeFlags(kwCampNegDeleteCmd, &kwCampNegDeleteFlags, true, false)
	addCampaignNegativeFlags(kwCampNegFindCmd, &kwCampNegFindFlags, true, false)
	kwCampNegEnableCmd.Flags().Int64Var(&kwCampNegEnableFlags.OrgID, "org-id", 0, "Organization ID override")
	kwCampNegEnableCmd.Flags().Int64Var(&kwCampNegEnableFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwCampNegEnableCmd.Flags().BoolVar(&kwCampNegEnableFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	kwCampNegEnableCmd.Flags().BoolVar(&kwCampNegEnableFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = kwCampNegEnableCmd.MarkFlagRequired("campaign-id")
	kwCampNegPauseCmd.Flags().Int64Var(&kwCampNegPauseFlags.OrgID, "org-id", 0, "Organization ID override")
	kwCampNegPauseCmd.Flags().Int64Var(&kwCampNegPauseFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	kwCampNegPauseCmd.Flags().BoolVar(&kwCampNegPauseFlags.DryRun, "dry-run", false, "Print payload without sending update request")
	kwCampNegPauseCmd.Flags().BoolVar(&kwCampNegPauseFlags.Yes, "yes", false, "Skip interactive confirmation")
	_ = kwCampNegPauseCmd.MarkFlagRequired("campaign-id")

	addAdGroupNegativeFlags(kwAdgNegListCmd, &kwAdgNegListFlags, false, true)
	addAdGroupNegativeFlags(kwAdgNegGetCmd, &kwAdgNegGetFlags, false, false)
	addAdGroupNegativeFlags(kwAdgNegCreateCmd, &kwAdgNegCreateFlags, true, false)
	addAdGroupNegativeFlags(kwAdgNegUpdateCmd, &kwAdgNegUpdateFlags, true, false)
	addAdGroupNegativeFlags(kwAdgNegDeleteCmd, &kwAdgNegDeleteFlags, true, false)
	addAdGroupNegativeFlags(kwAdgNegFindCmd, &kwAdgNegFindFlags, true, false)
}

func addCampaignNegativeFlags(cmd *cobra.Command, flags *campaignNegativeFlags, withBody, withAll bool) {
	cmd.Flags().Int64Var(&flags.OrgID, "org-id", 0, "Organization ID override")
	cmd.Flags().Int64Var(&flags.CampaignID, "campaign-id", 0, "Campaign ID")
	cmd.Flags().IntVar(&flags.Offset, "offset", 0, "Pagination offset")
	cmd.Flags().IntVar(&flags.Limit, "limit", 20, "Pagination limit")
	if withAll {
		cmd.Flags().BoolVar(&flags.All, "all", false, "Fetch all pages")
	}
	if withBody {
		cmd.Flags().StringVar(&flags.Body, "body", "", "Inline JSON body")
		cmd.Flags().StringVar(&flags.BodyFile, "body-file", "", "Path to JSON body file")
	}
	_ = cmd.MarkFlagRequired("campaign-id")
}

func addAdGroupNegativeFlags(cmd *cobra.Command, flags *adGroupNegativeFlags, withBody, withAll bool) {
	cmd.Flags().Int64Var(&flags.OrgID, "org-id", 0, "Organization ID override")
	cmd.Flags().Int64Var(&flags.CampaignID, "campaign-id", 0, "Campaign ID")
	cmd.Flags().Int64Var(&flags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	cmd.Flags().IntVar(&flags.Offset, "offset", 0, "Pagination offset")
	cmd.Flags().IntVar(&flags.Limit, "limit", 20, "Pagination limit")
	if withAll {
		cmd.Flags().BoolVar(&flags.All, "all", false, "Fetch all pages")
	}
	if withBody {
		cmd.Flags().StringVar(&flags.Body, "body", "", "Inline JSON body")
		cmd.Flags().StringVar(&flags.BodyFile, "body-file", "", "Path to JSON body file")
	}
	_ = cmd.MarkFlagRequired("campaign-id")
	_ = cmd.MarkFlagRequired("adgroup-id")
}

var kwCampNegListCmd = &cobra.Command{
	Use:   "list",
	Short: "List campaign negative keywords",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegListFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwCampNegListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if kwCampNegListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegListFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords", kwCampNegListFlags.CampaignID)
		q := url.Values{}
		q.Set("offset", strconv.Itoa(kwCampNegListFlags.Offset))
		q.Set("limit", strconv.Itoa(kwCampNegListFlags.Limit))
		return callListEndpoint(context.Background(), client, path, q, kwCampNegListFlags.Offset, kwCampNegListFlags.Limit, kwCampNegListFlags.All)
	},
}

var kwCampNegGetCmd = &cobra.Command{
	Use:   "get <keyword-id>",
	Short: "Get campaign negative keyword by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegGetFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		keywordID, err := parsePositiveInt64("keyword-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/%d", kwCampNegGetFlags.CampaignID, keywordID)
		return callAPIAndPrint(context.Background(), client, http.MethodGet, path, nil, nil)
	},
}

var kwCampNegCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create campaign negative keywords (bulk)",
	Example: "  appleads keywords campaign-negative create --campaign-id 123456 --body-file ./payloads/kw-cneg-create.json\n" +
		"  appleads keywords campaign-negative create --campaign-id 123456 --body '[{\"text\":\"free\",\"matchType\":\"EXACT\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegCreateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		payload, err := readJSONPayload(kwCampNegCreateFlags.Body, kwCampNegCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/bulk", kwCampNegCreateFlags.CampaignID)
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}

var kwCampNegUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update campaign negative keywords (bulk)",
	Example: "  appleads keywords campaign-negative update --campaign-id 123456 --body-file ./payloads/kw-cneg-update.json\n" +
		"  appleads keywords campaign-negative update --campaign-id 123456 --body '[{\"id\":111111,\"status\":\"PAUSED\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegUpdateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		payload, err := readJSONPayload(kwCampNegUpdateFlags.Body, kwCampNegUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/bulk", kwCampNegUpdateFlags.CampaignID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
	},
}

var kwCampNegDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete campaign negative keywords (bulk)",
	Example: "  appleads keywords campaign-negative delete --campaign-id 123456 --body-file ./payloads/kw-cneg-delete.json\n" +
		"  appleads keywords campaign-negative delete --campaign-id 123456 --body '[{\"id\":111111}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegDeleteFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		payload, err := readJSONPayload(kwCampNegDeleteFlags.Body, kwCampNegDeleteFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/bulk", kwCampNegDeleteFlags.CampaignID)
		return callAPIAndPrint(context.Background(), client, http.MethodDelete, path, nil, payload)
	},
}

var kwCampNegFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find campaign negative keywords by selector",
	Example: "  appleads keywords campaign-negative find --campaign-id 123456 --body-file ./payloads/kw-cneg-find.json\n" +
		"  appleads keywords campaign-negative find --campaign-id 123456 --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegFindFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		payload, err := readJSONPayload(kwCampNegFindFlags.Body, kwCampNegFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwCampNegFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/find", kwCampNegFindFlags.CampaignID)
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}

var kwCampNegEnableCmd = &cobra.Command{
	Use:   "enable <keyword-id>",
	Short: "Quick action: set campaign negative keyword status to ACTIVE",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegEnableFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
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
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/bulk", kwCampNegEnableFlags.CampaignID)
		return runSimpleStatusMutation(kwCampNegEnableFlags.OrgID, path, payload, kwCampNegEnableFlags.DryRun, kwCampNegEnableFlags.Yes, "campaign negative keyword enable")
	},
}

var kwCampNegPauseCmd = &cobra.Command{
	Use:   "pause <keyword-id>",
	Short: "Quick action: set campaign negative keyword status to PAUSED",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwCampNegPauseFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
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
		path := fmt.Sprintf("/campaigns/%d/negativekeywords/bulk", kwCampNegPauseFlags.CampaignID)
		return runSimpleStatusMutation(kwCampNegPauseFlags.OrgID, path, payload, kwCampNegPauseFlags.DryRun, kwCampNegPauseFlags.Yes, "campaign negative keyword pause")
	},
}

var kwAdgNegListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ad-group negative keywords",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegListFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegListFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		if kwAdgNegListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if kwAdgNegListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegListFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords", kwAdgNegListFlags.CampaignID, kwAdgNegListFlags.AdGroupID)
		q := url.Values{}
		q.Set("offset", strconv.Itoa(kwAdgNegListFlags.Offset))
		q.Set("limit", strconv.Itoa(kwAdgNegListFlags.Limit))
		return callListEndpoint(context.Background(), client, path, q, kwAdgNegListFlags.Offset, kwAdgNegListFlags.Limit, kwAdgNegListFlags.All)
	},
}

var kwAdgNegGetCmd = &cobra.Command{
	Use:   "get <keyword-id>",
	Short: "Get ad-group negative keyword by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegGetFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegGetFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		keywordID, err := parsePositiveInt64("keyword-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords/%d", kwAdgNegGetFlags.CampaignID, kwAdgNegGetFlags.AdGroupID, keywordID)
		return callAPIAndPrint(context.Background(), client, http.MethodGet, path, nil, nil)
	},
}

var kwAdgNegCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ad-group negative keywords (bulk)",
	Example: "  appleads keywords adgroup-negative create --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-aneg-create.json\n" +
		"  appleads keywords adgroup-negative create --campaign-id 123456 --adgroup-id 987654 --body '[{\"text\":\"free\",\"matchType\":\"EXACT\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegCreateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegCreateFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwAdgNegCreateFlags.Body, kwAdgNegCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords/bulk", kwAdgNegCreateFlags.CampaignID, kwAdgNegCreateFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}

var kwAdgNegUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ad-group negative keywords (bulk)",
	Example: "  appleads keywords adgroup-negative update --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-aneg-update.json\n" +
		"  appleads keywords adgroup-negative update --campaign-id 123456 --adgroup-id 987654 --body '[{\"id\":111111,\"status\":\"PAUSED\"}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegUpdateFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegUpdateFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwAdgNegUpdateFlags.Body, kwAdgNegUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords/bulk", kwAdgNegUpdateFlags.CampaignID, kwAdgNegUpdateFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPut, path, nil, payload)
	},
}

var kwAdgNegDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete ad-group negative keywords (bulk)",
	Example: "  appleads keywords adgroup-negative delete --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-aneg-delete.json\n" +
		"  appleads keywords adgroup-negative delete --campaign-id 123456 --adgroup-id 987654 --body '[{\"id\":111111}]'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegDeleteFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegDeleteFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwAdgNegDeleteFlags.Body, kwAdgNegDeleteFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords/bulk", kwAdgNegDeleteFlags.CampaignID, kwAdgNegDeleteFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodDelete, path, nil, payload)
	},
}

var kwAdgNegFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find ad-group negative keywords by selector",
	Example: "  appleads keywords adgroup-negative find --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/kw-aneg-find.json\n" +
		"  appleads keywords adgroup-negative find --campaign-id 123456 --adgroup-id 987654 --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kwAdgNegFindFlags.CampaignID <= 0 {
			return fmt.Errorf("--campaign-id must be > 0")
		}
		if kwAdgNegFindFlags.AdGroupID <= 0 {
			return fmt.Errorf("--adgroup-id must be > 0")
		}
		payload, err := readJSONPayload(kwAdgNegFindFlags.Body, kwAdgNegFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), kwAdgNegFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/campaigns/%d/adgroups/%d/negativekeywords/find", kwAdgNegFindFlags.CampaignID, kwAdgNegFindFlags.AdGroupID)
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}
