package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var adsListFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Offset     int
	Limit      int
	All        bool
}

var adsGetFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
}

var adsCreateFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

var adsUpdateFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
	Body       string
	BodyFile   string
}

var adsDeleteFlags struct {
	OrgID      int64
	CampaignID int64
	AdGroupID  int64
}

var adsFindFlags struct {
	OrgID      int64
	CampaignID int64
	Body       string
	BodyFile   string
}

func init() {
	adsCmd.AddCommand(adsListCmd)
	adsCmd.AddCommand(adsGetCmd)
	adsCmd.AddCommand(adsCreateCmd)
	adsCmd.AddCommand(adsUpdateCmd)
	adsCmd.AddCommand(adsDeleteCmd)
	adsCmd.AddCommand(adsFindCmd)

	adsListCmd.Flags().Int64Var(&adsListFlags.OrgID, "org-id", 0, "Organization ID override")
	adsListCmd.Flags().Int64Var(&adsListFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsListCmd.Flags().Int64Var(&adsListFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	adsListCmd.Flags().IntVar(&adsListFlags.Offset, "offset", 0, "Pagination offset")
	adsListCmd.Flags().IntVar(&adsListFlags.Limit, "limit", 20, "Pagination limit")
	adsListCmd.Flags().BoolVar(&adsListFlags.All, "all", false, "Fetch all pages")

	adsGetCmd.Flags().Int64Var(&adsGetFlags.OrgID, "org-id", 0, "Organization ID override")
	adsGetCmd.Flags().Int64Var(&adsGetFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsGetCmd.Flags().Int64Var(&adsGetFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")

	adsCreateCmd.Flags().Int64Var(&adsCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	adsCreateCmd.Flags().Int64Var(&adsCreateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsCreateCmd.Flags().Int64Var(&adsCreateFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	adsCreateCmd.Flags().StringVar(&adsCreateFlags.Body, "body", "", "Inline JSON body")
	adsCreateCmd.Flags().StringVar(&adsCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	adsUpdateCmd.Flags().Int64Var(&adsUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	adsUpdateCmd.Flags().Int64Var(&adsUpdateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsUpdateCmd.Flags().Int64Var(&adsUpdateFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")
	adsUpdateCmd.Flags().StringVar(&adsUpdateFlags.Body, "body", "", "Inline JSON body")
	adsUpdateCmd.Flags().StringVar(&adsUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	adsDeleteCmd.Flags().Int64Var(&adsDeleteFlags.OrgID, "org-id", 0, "Organization ID override")
	adsDeleteCmd.Flags().Int64Var(&adsDeleteFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adsDeleteCmd.Flags().Int64Var(&adsDeleteFlags.AdGroupID, "adgroup-id", 0, "Ad group ID")

	adsFindCmd.Flags().Int64Var(&adsFindFlags.OrgID, "org-id", 0, "Organization ID override")
	adsFindCmd.Flags().Int64Var(&adsFindFlags.CampaignID, "campaign-id", 0, "Campaign ID (optional; falls back to org-level find)")
	adsFindCmd.Flags().StringVar(&adsFindFlags.Body, "body", "", "Inline JSON selector body")
	adsFindCmd.Flags().StringVar(&adsFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
}

var adsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ads (org/campaign/ad-group scope based on flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adsListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if adsListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}

		client, _, _, err := authedClient(context.Background(), adsListFlags.OrgID, true)
		if err != nil {
			return err
		}

		path := "/ads"
		if adsListFlags.CampaignID > 0 && adsListFlags.AdGroupID > 0 {
			path = fmt.Sprintf("/campaigns/%d/adgroups/%d/ads", adsListFlags.CampaignID, adsListFlags.AdGroupID)
		} else if adsListFlags.CampaignID > 0 {
			path = fmt.Sprintf("/campaigns/%d/ads", adsListFlags.CampaignID)
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(adsListFlags.Offset))
		q.Set("limit", strconv.Itoa(adsListFlags.Limit))
		err = callListEndpoint(context.Background(), client, path, q, adsListFlags.Offset, adsListFlags.Limit, adsListFlags.All)
		if adsListFlags.CampaignID == 0 && adsListFlags.AdGroupID == 0 {
			return withNotFoundHint(err, "org-level ads endpoint unavailable for this account, try with --campaign-id and --adgroup-id")
		}
		if adsListFlags.CampaignID > 0 && adsListFlags.AdGroupID == 0 {
			return withNotFoundHint(err, "campaign-level ads endpoint unavailable for this account, try with --adgroup-id")
		}
		return err
	},
}

var adsGetCmd = &cobra.Command{
	Use:   "get <ad-id>",
	Short: "Get ad by ID (scoped path is used when flags are provided)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adID, err := parsePositiveInt64("ad-id", args[0])
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), adsGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{}
		if adsGetFlags.CampaignID > 0 && adsGetFlags.AdGroupID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/adgroups/%d/ads/%d", adsGetFlags.CampaignID, adsGetFlags.AdGroupID, adID))
		}
		if adsGetFlags.CampaignID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/ads/%d", adsGetFlags.CampaignID, adID))
		}
		paths = append(paths, fmt.Sprintf("/ads/%d", adID))
		if len(paths) == 1 {
			return callAPIAndPrint(context.Background(), client, http.MethodGet, paths[0], nil, nil)
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodGet, paths, nil, nil)
	},
}

var adsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ad with JSON payload (org/campaign/ad-group scope based on flags)",
	Example: "  appleads ads create --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/ad-create.json\n" +
		"  appleads ads create --campaign-id 123456 --adgroup-id 987654 --body '{\"status\":\"ENABLED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(adsCreateFlags.Body, adsCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), adsCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		if adsCreateFlags.CampaignID > 0 && adsCreateFlags.AdGroupID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
				fmt.Sprintf("/campaigns/%d/adgroups/%d/ads", adsCreateFlags.CampaignID, adsCreateFlags.AdGroupID),
				fmt.Sprintf("/campaigns/%d/ads", adsCreateFlags.CampaignID),
				"/ads",
			}, nil, payload)
		}
		if adsCreateFlags.CampaignID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
				fmt.Sprintf("/campaigns/%d/ads", adsCreateFlags.CampaignID),
				"/ads",
			}, nil, payload)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/ads", nil, payload)
	},
}

var adsUpdateCmd = &cobra.Command{
	Use:   "update <ad-id>",
	Short: "Update ad by ID with JSON payload (scoped path is used when flags are provided)",
	Args:  cobra.ExactArgs(1),
	Example: "  appleads ads update 333333 --campaign-id 123456 --adgroup-id 987654 --body-file ./payloads/ad-update.json\n" +
		"  appleads ads update 333333 --campaign-id 123456 --adgroup-id 987654 --body '{\"status\":\"PAUSED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		adID, err := parsePositiveInt64("ad-id", args[0])
		if err != nil {
			return err
		}

		payload, err := readJSONPayload(adsUpdateFlags.Body, adsUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), adsUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{}
		if adsUpdateFlags.CampaignID > 0 && adsUpdateFlags.AdGroupID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/adgroups/%d/ads/%d", adsUpdateFlags.CampaignID, adsUpdateFlags.AdGroupID, adID))
		}
		if adsUpdateFlags.CampaignID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/ads/%d", adsUpdateFlags.CampaignID, adID))
		}
		paths = append(paths, fmt.Sprintf("/ads/%d", adID))
		if len(paths) == 1 {
			return callAPIAndPrint(context.Background(), client, http.MethodPut, paths[0], nil, payload)
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPut, paths, nil, payload)
	},
}

var adsDeleteCmd = &cobra.Command{
	Use:   "delete <ad-id>",
	Short: "Delete ad by ID (scoped path is used when flags are provided)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adID, err := parsePositiveInt64("ad-id", args[0])
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), adsDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		paths := []string{}
		if adsDeleteFlags.CampaignID > 0 && adsDeleteFlags.AdGroupID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/adgroups/%d/ads/%d", adsDeleteFlags.CampaignID, adsDeleteFlags.AdGroupID, adID))
		}
		if adsDeleteFlags.CampaignID > 0 {
			paths = append(paths, fmt.Sprintf("/campaigns/%d/ads/%d", adsDeleteFlags.CampaignID, adID))
		}
		paths = append(paths, fmt.Sprintf("/ads/%d", adID))
		if len(paths) == 1 {
			return callAPIAndPrint(context.Background(), client, http.MethodDelete, paths[0], nil, nil)
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodDelete, paths, nil, nil)
	},
}

var adsFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find ads by selector (campaign-level or org-level)",
	Example: "  appleads ads find --body-file ./payloads/ad-find.json\n" +
		"  appleads ads find --campaign-id 123456 --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(adsFindFlags.Body, adsFindFlags.BodyFile, false)
		if err != nil {
			return err
		}

		client, _, _, err := authedClient(context.Background(), adsFindFlags.OrgID, true)
		if err != nil {
			return err
		}

		path := "/ads/find"
		if adsFindFlags.CampaignID > 0 {
			path = fmt.Sprintf("/campaigns/%d/ads/find", adsFindFlags.CampaignID)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, path, nil, payload)
	},
}
