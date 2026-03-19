package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var adGroupsListFlags struct {
	OrgID      int64
	CampaignID int64
	Offset     int
	Limit      int
	All        bool
}

var adGroupsGetFlags struct {
	OrgID      int64
	CampaignID int64
}

var adGroupsCreateFlags struct {
	OrgID      int64
	CampaignID int64
	Body       string
	BodyFile   string
}

var adGroupsUpdateFlags struct {
	OrgID      int64
	CampaignID int64
	Body       string
	BodyFile   string
}

var adGroupsDeleteFlags struct {
	OrgID      int64
	CampaignID int64
}

var adGroupsFindFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

func init() {
	adGroupsCmd.AddCommand(adGroupsListCmd)
	adGroupsCmd.AddCommand(adGroupsGetCmd)
	adGroupsCmd.AddCommand(adGroupsCreateCmd)
	adGroupsCmd.AddCommand(adGroupsUpdateCmd)
	adGroupsCmd.AddCommand(adGroupsDeleteCmd)
	adGroupsCmd.AddCommand(adGroupsFindCmd)

	adGroupsListCmd.Flags().Int64Var(&adGroupsListFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsListCmd.Flags().Int64Var(&adGroupsListFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adGroupsListCmd.Flags().IntVar(&adGroupsListFlags.Offset, "offset", 0, "Pagination offset")
	adGroupsListCmd.Flags().IntVar(&adGroupsListFlags.Limit, "limit", 20, "Pagination limit")
	adGroupsListCmd.Flags().BoolVar(&adGroupsListFlags.All, "all", false, "Fetch all pages")

	adGroupsGetCmd.Flags().Int64Var(&adGroupsGetFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsGetCmd.Flags().Int64Var(&adGroupsGetFlags.CampaignID, "campaign-id", 0, "Campaign ID")

	adGroupsCreateCmd.Flags().Int64Var(&adGroupsCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsCreateCmd.Flags().Int64Var(&adGroupsCreateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adGroupsCreateCmd.Flags().StringVar(&adGroupsCreateFlags.Body, "body", "", "Inline JSON body")
	adGroupsCreateCmd.Flags().StringVar(&adGroupsCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	adGroupsUpdateCmd.Flags().Int64Var(&adGroupsUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsUpdateCmd.Flags().Int64Var(&adGroupsUpdateFlags.CampaignID, "campaign-id", 0, "Campaign ID")
	adGroupsUpdateCmd.Flags().StringVar(&adGroupsUpdateFlags.Body, "body", "", "Inline JSON body")
	adGroupsUpdateCmd.Flags().StringVar(&adGroupsUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	adGroupsDeleteCmd.Flags().Int64Var(&adGroupsDeleteFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsDeleteCmd.Flags().Int64Var(&adGroupsDeleteFlags.CampaignID, "campaign-id", 0, "Campaign ID")

	adGroupsFindCmd.Flags().Int64Var(&adGroupsFindFlags.OrgID, "org-id", 0, "Organization ID override")
	adGroupsFindCmd.Flags().StringVar(&adGroupsFindFlags.Body, "body", "", "Inline JSON selector body")
	adGroupsFindCmd.Flags().StringVar(&adGroupsFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
}

var adGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ad groups (campaign-level when --campaign-id is provided)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adGroupsListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if adGroupsListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}

		client, _, _, err := authedClient(context.Background(), adGroupsListFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(adGroupsListFlags.Offset))
		q.Set("limit", strconv.Itoa(adGroupsListFlags.Limit))
		path := "/adgroups"
		if adGroupsListFlags.CampaignID > 0 {
			path = fmt.Sprintf("/campaigns/%d/adgroups", adGroupsListFlags.CampaignID)
		}
		err = callListEndpoint(context.Background(), client, path, q, adGroupsListFlags.Offset, adGroupsListFlags.Limit, adGroupsListFlags.All)
		if adGroupsListFlags.CampaignID == 0 {
			return withNotFoundHint(err, "org-level adgroups endpoint unavailable for this account, try again with --campaign-id")
		}
		return err
	},
}

var adGroupsGetCmd = &cobra.Command{
	Use:   "get <adgroup-id>",
	Short: "Get ad group by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adGroupID, err := parsePositiveInt64("adgroup-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), adGroupsGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		if adGroupsGetFlags.CampaignID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodGet, []string{
				fmt.Sprintf("/campaigns/%d/adgroups/%d", adGroupsGetFlags.CampaignID, adGroupID),
				fmt.Sprintf("/adgroups/%d", adGroupID),
			}, nil, nil)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodGet, fmt.Sprintf("/adgroups/%d", adGroupID), nil, nil)
	},
}

var adGroupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ad group with JSON payload (campaign-level when --campaign-id is provided)",
	Example: "  appleads adgroups create --campaign-id 123456 --body-file ./payloads/adgroup-create.json\n" +
		"  appleads adgroups create --campaign-id 123456 --body '{\"name\":\"AG-1\",\"status\":\"ENABLED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(adGroupsCreateFlags.Body, adGroupsCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), adGroupsCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		if adGroupsCreateFlags.CampaignID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
				fmt.Sprintf("/campaigns/%d/adgroups", adGroupsCreateFlags.CampaignID),
				"/adgroups",
			}, nil, payload)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/adgroups", nil, payload)
	},
}

var adGroupsUpdateCmd = &cobra.Command{
	Use:   "update <adgroup-id>",
	Short: "Update ad group by ID with JSON payload",
	Args:  cobra.ExactArgs(1),
	Example: "  appleads adgroups update 987654 --campaign-id 123456 --body-file ./payloads/adgroup-update.json\n" +
		"  appleads adgroups update 987654 --campaign-id 123456 --body '{\"status\":\"PAUSED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		adGroupID, err := parsePositiveInt64("adgroup-id", args[0])
		if err != nil {
			return err
		}
		payload, err := readJSONPayload(adGroupsUpdateFlags.Body, adGroupsUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), adGroupsUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		if adGroupsUpdateFlags.CampaignID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPut, []string{
				fmt.Sprintf("/campaigns/%d/adgroups/%d", adGroupsUpdateFlags.CampaignID, adGroupID),
				fmt.Sprintf("/adgroups/%d", adGroupID),
			}, nil, payload)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPut, fmt.Sprintf("/adgroups/%d", adGroupID), nil, payload)
	},
}

var adGroupsDeleteCmd = &cobra.Command{
	Use:   "delete <adgroup-id>",
	Short: "Delete ad group by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adGroupID, err := parsePositiveInt64("adgroup-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), adGroupsDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		if adGroupsDeleteFlags.CampaignID > 0 {
			return callAPIAndPrintWithFallback(context.Background(), client, http.MethodDelete, []string{
				fmt.Sprintf("/campaigns/%d/adgroups/%d", adGroupsDeleteFlags.CampaignID, adGroupID),
				fmt.Sprintf("/adgroups/%d", adGroupID),
			}, nil, nil)
		}
		return callAPIAndPrint(context.Background(), client, http.MethodDelete, fmt.Sprintf("/adgroups/%d", adGroupID), nil, nil)
	},
}

var adGroupsFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find ad groups with selector JSON payload",
	Example: "  appleads adgroups find --body-file ./payloads/adgroup-find.json\n" +
		"  appleads adgroups find --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(adGroupsFindFlags.Body, adGroupsFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), adGroupsFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/adgroups/find", nil, payload)
	},
}
