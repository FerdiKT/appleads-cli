package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var campaignsCreateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var campaignsUpdateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var campaignsDeleteFlags struct {
	OrgID int64
}

var campaignsGetFlags struct {
	OrgID int64
}

var campaignsFindFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

func init() {
	campaignsCmd.AddCommand(campaignsGetCmd)
	campaignsCmd.AddCommand(campaignsCreateCmd)
	campaignsCmd.AddCommand(campaignsUpdateCmd)
	campaignsCmd.AddCommand(campaignsDeleteCmd)
	campaignsCmd.AddCommand(campaignsFindCmd)

	campaignsGetCmd.Flags().Int64Var(&campaignsGetFlags.OrgID, "org-id", 0, "Organization ID override")

	campaignsCreateCmd.Flags().Int64Var(&campaignsCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	campaignsCreateCmd.Flags().StringVar(&campaignsCreateFlags.Body, "body", "", "Inline JSON body")
	campaignsCreateCmd.Flags().StringVar(&campaignsCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	campaignsUpdateCmd.Flags().Int64Var(&campaignsUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	campaignsUpdateCmd.Flags().StringVar(&campaignsUpdateFlags.Body, "body", "", "Inline JSON body")
	campaignsUpdateCmd.Flags().StringVar(&campaignsUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	campaignsDeleteCmd.Flags().Int64Var(&campaignsDeleteFlags.OrgID, "org-id", 0, "Organization ID override")

	campaignsFindCmd.Flags().Int64Var(&campaignsFindFlags.OrgID, "org-id", 0, "Organization ID override")
	campaignsFindCmd.Flags().StringVar(&campaignsFindFlags.Body, "body", "", "Inline JSON selector body")
	campaignsFindCmd.Flags().StringVar(&campaignsFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
}

var campaignsGetCmd = &cobra.Command{
	Use:   "get <campaign-id>",
	Short: "Get campaign by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		campaignID, err := parsePositiveInt64("campaign-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), campaignsGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodGet, fmt.Sprintf("/campaigns/%d", campaignID), nil, nil)
	},
}

var campaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create campaign with JSON payload",
	Example: "  appleads campaigns create --body-file ./payloads/campaign-create.json\n" +
		"  appleads campaigns create --body '{\"name\":\"My Campaign\",\"status\":\"ENABLED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(campaignsCreateFlags.Body, campaignsCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), campaignsCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/campaigns", nil, payload)
	},
}

var campaignsUpdateCmd = &cobra.Command{
	Use:   "update <campaign-id>",
	Short: "Update campaign by ID with JSON payload",
	Args:  cobra.ExactArgs(1),
	Example: "  appleads campaigns update 123456 --body-file ./payloads/campaign-update.json\n" +
		"  appleads campaigns update 123456 --body '{\"status\":\"PAUSED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		campaignID, err := parsePositiveInt64("campaign-id", args[0])
		if err != nil {
			return err
		}
		payload, err := readJSONPayload(campaignsUpdateFlags.Body, campaignsUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), campaignsUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPut, fmt.Sprintf("/campaigns/%d", campaignID), nil, payload)
	},
}

var campaignsDeleteCmd = &cobra.Command{
	Use:   "delete <campaign-id>",
	Short: "Delete campaign by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		campaignID, err := parsePositiveInt64("campaign-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), campaignsDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodDelete, fmt.Sprintf("/campaigns/%d", campaignID), nil, nil)
	},
}

var campaignsFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find campaigns with selector JSON payload",
	Example: "  appleads campaigns find --body-file ./payloads/campaign-find.json\n" +
		"  appleads campaigns find --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(campaignsFindFlags.Body, campaignsFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), campaignsFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/campaigns/find", nil, payload)
	},
}
