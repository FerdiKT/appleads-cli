package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var creativeListFlags struct {
	OrgID  int64
	Offset int
	Limit  int
	All    bool
}

var creativeGetFlags struct {
	OrgID int64
}

var creativeFindFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var creativeCreateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var creativeUpdateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var creativeDeleteFlags struct {
	OrgID int64
}

func init() {
	creativesCmd.AddCommand(creativeListCmd)
	creativesCmd.AddCommand(creativeGetCmd)
	creativesCmd.AddCommand(creativeFindCmd)
	creativesCmd.AddCommand(creativeCreateCmd)
	creativesCmd.AddCommand(creativeUpdateCmd)
	creativesCmd.AddCommand(creativeDeleteCmd)

	creativeListCmd.Flags().Int64Var(&creativeListFlags.OrgID, "org-id", 0, "Organization ID override")
	creativeListCmd.Flags().IntVar(&creativeListFlags.Offset, "offset", 0, "Pagination offset")
	creativeListCmd.Flags().IntVar(&creativeListFlags.Limit, "limit", 20, "Pagination limit")
	creativeListCmd.Flags().BoolVar(&creativeListFlags.All, "all", false, "Fetch all pages")

	creativeGetCmd.Flags().Int64Var(&creativeGetFlags.OrgID, "org-id", 0, "Organization ID override")

	creativeFindCmd.Flags().Int64Var(&creativeFindFlags.OrgID, "org-id", 0, "Organization ID override")
	creativeFindCmd.Flags().StringVar(&creativeFindFlags.Body, "body", "", "Inline JSON selector body")
	creativeFindCmd.Flags().StringVar(&creativeFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")

	creativeCreateCmd.Flags().Int64Var(&creativeCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	creativeCreateCmd.Flags().StringVar(&creativeCreateFlags.Body, "body", "", "Inline JSON body")
	creativeCreateCmd.Flags().StringVar(&creativeCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	creativeUpdateCmd.Flags().Int64Var(&creativeUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	creativeUpdateCmd.Flags().StringVar(&creativeUpdateFlags.Body, "body", "", "Inline JSON body")
	creativeUpdateCmd.Flags().StringVar(&creativeUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	creativeDeleteCmd.Flags().Int64Var(&creativeDeleteFlags.OrgID, "org-id", 0, "Organization ID override")
}

var creativeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List creatives",
	RunE: func(cmd *cobra.Command, args []string) error {
		if creativeListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if creativeListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), creativeListFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(creativeListFlags.Offset))
		q.Set("limit", strconv.Itoa(creativeListFlags.Limit))
		return callListEndpoint(context.Background(), client, "/creatives", q, creativeListFlags.Offset, creativeListFlags.Limit, creativeListFlags.All)
	},
}

var creativeGetCmd = &cobra.Command{
	Use:   "get <creative-id>",
	Short: "Get creative by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creativeID, err := parsePositiveInt64("creative-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), creativeGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodGet, fmt.Sprintf("/creatives/%d", creativeID), nil, nil)
	},
}

var creativeFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find creatives with selector payload",
	Example: "  appleads creatives find --body-file ./payloads/creative-find.json\n" +
		"  appleads creatives find --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(creativeFindFlags.Body, creativeFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), creativeFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodPost, "/creatives/find", nil, payload)
	},
}

var creativeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create creative with JSON payload",
	Example: "  appleads creatives create --body-file ./payloads/creative-create.json\n" +
		"  appleads creatives create --body '{\"name\":\"Creative Set 1\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(creativeCreateFlags.Body, creativeCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), creativeCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
			"/creatives",
			"/creative-sets",
		}, nil, payload)
	},
}

var creativeUpdateCmd = &cobra.Command{
	Use:   "update <creative-id>",
	Short: "Update creative by ID with JSON payload",
	Args:  cobra.ExactArgs(1),
	Example: "  appleads creatives update 123456 --body-file ./payloads/creative-update.json\n" +
		"  appleads creatives update 123456 --body '{\"status\":\"PAUSED\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		creativeID, err := parsePositiveInt64("creative-id", args[0])
		if err != nil {
			return err
		}
		payload, err := readJSONPayload(creativeUpdateFlags.Body, creativeUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), creativeUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPut, []string{
			fmt.Sprintf("/creatives/%d", creativeID),
			fmt.Sprintf("/creative-sets/%d", creativeID),
		}, nil, payload)
	},
}

var creativeDeleteCmd = &cobra.Command{
	Use:   "delete <creative-id>",
	Short: "Delete creative by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creativeID, err := parsePositiveInt64("creative-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), creativeDeleteFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodDelete, []string{
			fmt.Sprintf("/creatives/%d", creativeID),
			fmt.Sprintf("/creative-sets/%d", creativeID),
		}, nil, nil)
	},
}
